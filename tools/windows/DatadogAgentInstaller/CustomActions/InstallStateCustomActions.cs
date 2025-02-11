using System;
using Datadog.CustomActions.Extensions;
using Datadog.CustomActions.Interfaces;
using Datadog.CustomActions.Native;
using Microsoft.Deployment.WindowsInstaller;
using Microsoft.Win32;
using ServiceController = Datadog.CustomActions.Native.ServiceController;

namespace Datadog.CustomActions
{
    // Fetch and process registry value(s) and return a string to be assigned to a WIX property.
    using GetRegistryPropertyHandler = Func<string>;

    public class InstallStateCustomActions
    {
        private readonly ISession _session;
        private readonly IRegistryServices _registryServices;
        private readonly IServiceController _serviceController;

        public InstallStateCustomActions(
            ISession session,
            IRegistryServices registryServices,
            IServiceController serviceController)
        {
            _session = session;
            _registryServices = registryServices;
            _serviceController = serviceController;
        }

        public InstallStateCustomActions(ISession session)
            : this(
                session,
                new RegistryServices(),
                new ServiceController())
        {
        }

        /// <summary>
        /// If the WIX property <c>propertyName</c> does not have a value, assign it the value returned by <c>handler</c>.
        /// This gives precedence to properties provided on the command line over the registry values.
        /// </summary>
        private static void RegistryProperty(ISession session, string propertyName, GetRegistryPropertyHandler handler)
        {
            if (string.IsNullOrEmpty(session[propertyName]))
            {
                try
                {
                    var propertyVal = handler();
                    if (!string.IsNullOrEmpty(propertyVal))
                    {
                        session[propertyName] = propertyVal;
                        session.Log($"Found {propertyName} in registry {session[propertyName]}");
                    }
                }
                catch (Exception e)
                {
                    session.Log($"Exception processing registry value for {propertyName}: {e}");
                }
            }
            else
            {
                session.Log($"User provided {propertyName} {session[propertyName]}");
            }
        }

        /// <summary>
        /// Convenience wrapper of <c>RegistryProperty</c> for properties that have an exact 1:1 mapping to a registry value
        /// and don't require additional processing.
        /// </summary>
        private static void RegistryValueProperty(ISession session, string propertyName, IRegistryKey registryKey,
            string registryValue)
        {
            RegistryProperty(session, propertyName,
                () => registryKey.GetValue(registryValue)?.ToString());
        }

        /// <summary>
        /// Assigns WIX properties that were not provided by the user to their registry values.
        /// </summary>
        /// <remarks>
        /// Custom Action that runs (only once) in either the InstallUISequence or the InstallExecuteSequence.
        ///
        /// During removing-for-upgrade the installer being removed does not receive any properties from the
        /// installer being installed, only UPGRADINGPRODUCTCODE is set. Thus the state for the installer being
        /// removed will come from the registry values only.
        /// </remarks>
        public ActionResult ReadInstallState()
        {
            try
            {
                using var subkey =
                    _registryServices.OpenRegistryKey(Registries.LocalMachine, Constants.DatadogAgentRegistryKey);
                if (subkey != null)
                {
                    // DDAGENTUSER_NAME
                    //
                    // The user account can be provided to the installer by
                    // * The registry
                    // * The command line
                    // * The agent user dialog
                    // The user account domain and name are stored separately in the registry
                    // but are passed together on the command line and the agent user dialog.
                    // This function will combine the registry properties if they exist.
                    // Preference is given to creds provided on the command line and the agent user dialog.
                    // For UI installs it ensures that the agent user dialog is pre-populated.
                    RegistryProperty(_session, "DDAGENTUSER_NAME",
                        () =>
                        {
                            var domain = subkey.GetValue("installedDomain")?.ToString();
                            var user = subkey.GetValue("installedUser")?.ToString();
                            if (!string.IsNullOrEmpty(domain) && !string.IsNullOrEmpty(user))
                            {
                                return $"{domain}\\{user}";
                            }

                            return string.Empty;
                        });

                    RegistryValueProperty(_session, "PROJECTLOCATION", subkey, "InstallPath");
                    RegistryValueProperty(_session, "APPLICATIONDATADIRECTORY", subkey, "ConfigRoot");
                }

                GetWindowsBuildVersion();
            }
            catch (Exception e)
            {
                _session.Log($"Error reading install state: {e}");
                return ActionResult.Failure;
            }

            return ActionResult.Success;
        }

        /// <summary>
        /// WiX doesn't support getting the real build number on Windows 10+ so we must fetch it ourselves
        /// </summary>
        public void GetWindowsBuildVersion()
        {
            using var subkey = _registryServices.OpenRegistryKey(Registries.LocalMachine,
                @"Software\Microsoft\Windows NT\CurrentVersion");
            if (subkey != null)
            {
                var currentBuild = subkey.GetValue("CurrentBuild");
                if (currentBuild != null)
                {
                    _session["DDAGENT_WINDOWSBUILD"] = subkey.GetValue("CurrentBuild").ToString();
                    _session.Log($"WindowsBuild: {_session["DDAGENT_WINDOWSBUILD"]}");
                }
            }
            else
            {
                _session.Log("WindowsBuild not found");
            }
        }

        [CustomAction]
        public static ActionResult ReadInstallState(Session session)
        {
            return new InstallStateCustomActions(new SessionWrapper(session)).ReadInstallState();
        }

        /// <summary>
        /// Deferred custom action that stores properties in the registry
        /// </summary>
        /// <remarks>
        /// WiX RegistryValue elements are only written when their parent Feature is installed. This means
        /// that on change/modify operations the registry keys are not updated. This custom action writes
        /// the properties to the registry that we need to change during change/modify installer operations.
        /// </remarks>
        public ActionResult WriteInstallState()
        {
            try
            {
                using var subkey =
                    _registryServices.CreateRegistryKey(Registries.LocalMachine, Constants.DatadogAgentRegistryKey);
                if (subkey == null)
                {
                    throw new Exception("Unable to create agent registry key");
                }

                _session.Log($"Storing installedDomain={_session.Property("DDAGENTUSER_PROCESSED_DOMAIN")}");
                subkey.SetValue("installedDomain", _session.Property("DDAGENTUSER_PROCESSED_DOMAIN"),
                    RegistryValueKind.String);
                _session.Log($"Storing installedUser={_session.Property("DDAGENTUSER_PROCESSED_NAME")}");
                subkey.SetValue("installedUser", _session.Property("DDAGENTUSER_PROCESSED_NAME"),
                    RegistryValueKind.String);
            }
            catch (Exception e)
            {
                _session.Log($"Error storing registry properties: {e}");
                return ActionResult.Failure;
            }

            return ActionResult.Success;
        }

        [CustomAction]
        public static ActionResult WriteInstallState(Session session)
        {
            return new InstallStateCustomActions(new SessionWrapper(session)).WriteInstallState();
        }


        /// <summary>
        /// Uninstall CA that removes the changes from the WriteInstallState CA
        /// </summary>
        /// <remarks>
        /// If these registry values are not removed then MSI won't remove the key.
        /// </remarks>
        public ActionResult UninstallWriteInstallState()
        {
            try
            {
                using var subkey =
                    _registryServices.OpenRegistryKey(Registries.LocalMachine, Constants.DatadogAgentRegistryKey,
                        writable: true);
                if (subkey == null)
                {
                    // registry key does not exist, nothing to do
                    _session.Log(
                        $"Registry key HKLM\\{Constants.DatadogAgentRegistryKey} does not exist, there are no values to remove.");
                    return ActionResult.Success;
                }

                foreach (var value in new[]
                         {
                             "installedDomain",
                             "installedUser"
                         })
                {
                    try
                    {
                        subkey.DeleteValue(value);
                    }
                    catch (Exception e)
                    {
                        // Don't print stack trace as it may be seen as a terminal error by readers of the log.
                        _session.Log($"Warning, cannot removing registry value: {e.Message}");
                    }
                }
            }
            catch (Exception e)
            {
                _session.Log($"Warning, could not access registry key {Constants.DatadogAgentRegistryKey}: {e}");
                // This step can fail without failing the un-installation.
            }

            return ActionResult.Success;
        }

        [CustomAction]
        public static ActionResult UninstallWriteInstallState(Session session)
        {
            return new InstallStateCustomActions(new SessionWrapper(session)).UninstallWriteInstallState();
        }
    }
}
