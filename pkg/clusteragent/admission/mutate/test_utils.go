// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

package mutate

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeEnvWithValue(name, value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  name,
		Value: value,
	}
}

func fakeEnvWithFieldRefValue(name, value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: value,
			},
		},
	}
}

func fakeEnv(name string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  name,
		Value: name + "-env-value",
	}
}

func fakeContainer(name string) corev1.Container {
	return corev1.Container{
		Name: name,
		Env: []corev1.EnvVar{
			fakeEnv(name + "-env-foo"),
			fakeEnv(name + "-env-bar"),
		},
	}
}

func fakePodWithContainer(name string, containers ...corev1.Container) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PodSpec{
			Containers: containers,
		},
	}
}

func fakePodWithInitContainer(name string, containers ...corev1.Container) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PodSpec{
			InitContainers: containers,
		},
	}
}

func withLabels(pod *corev1.Pod, labels map[string]string) *corev1.Pod {
	pod.Labels = labels
	return pod
}

func fakePodWithLabel(k, v string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				k: v,
			},
		},
	}
}

func fakePodWithAnnotation(k, v string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod",
			Annotations: map[string]string{
				k: v,
			},
		},
	}
	return withContainer(pod, "-container")
}

func fakePodWithParent(ns string, as, ls map[string]string, es []corev1.EnvVar, parentKind, parentName string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod",
			Namespace:       ns,
			Annotations:     as,
			Labels:          ls,
			OwnerReferences: []metav1.OwnerReference{},
		},
	}
	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
		Name: pod.Name,
		Env:  es,
	})

	var ownerRef *metav1.OwnerReference
	objMeta := metav1.ObjectMeta{
		Name:      parentName,
		Namespace: ns,
	}

	if parentKind == "replicaset" {
		rs := &appsv1.ReplicaSet{
			ObjectMeta: objMeta,
		}
		ownerRef = metav1.NewControllerRef(rs, appsv1.SchemeGroupVersion.WithKind("ReplicaSet"))
	} else if parentKind == "statefulset" {
		ss := &appsv1.StatefulSet{
			ObjectMeta: objMeta,
		}
		ownerRef = metav1.NewControllerRef(ss, appsv1.SchemeGroupVersion.WithKind("StatefulSet"))
	} else if parentKind == "job" {
		j := &batchv1.Job{
			ObjectMeta: objMeta,
		}
		ownerRef = metav1.NewControllerRef(j, batchv1.SchemeGroupVersion.WithKind("Job"))
	} else if parentKind == "daemonset" {
		ds := &appsv1.DaemonSet{
			ObjectMeta: objMeta,
		}
		ownerRef = metav1.NewControllerRef(ds, appsv1.SchemeGroupVersion.WithKind("DaemonSet"))
	} else {
		return pod
	}
	pod.ObjectMeta.OwnerReferences = append(pod.ObjectMeta.OwnerReferences, *ownerRef)

	return pod
}

func fakePodWithAnnotations(as map[string]string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "pod",
			Annotations: as,
		},
	}
	return withContainer(pod, "-container")
}

func fakePodWithEnv(name, env string) *corev1.Pod {
	return fakePodWithContainer(name, corev1.Container{Name: name + "-container", Env: []corev1.EnvVar{fakeEnv(env)}})
}

func fakePodWithEnvValue(name, envKey, envVal string) *corev1.Pod {
	return fakePodWithContainer(name, corev1.Container{Name: name + "-container", Env: []corev1.EnvVar{fakeEnvWithValue(envKey, envVal)}})
}

func fakePodWithEnvFieldRefValue(name, envKey, path string) *corev1.Pod {
	return fakePodWithContainer(name, corev1.Container{Name: name + "-container", Env: []corev1.EnvVar{fakeEnvWithFieldRefValue(envKey, path)}})
}

func fakePodWithNamespaceAndLabel(namespace, k, v string) *corev1.Pod {
	pod := fakePodWithLabel(k, v)
	pod.Namespace = namespace

	return pod
}

func fakePodWithVolume(podName, volumeName, mountPath string) *corev1.Pod {
	pod := fakePod(podName)
	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{Name: volumeName})
	pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{Name: volumeName, MountPath: mountPath})
	return pod
}

func fakePod(name string) *corev1.Pod {
	return fakePodWithContainer(name, corev1.Container{Name: name + "-container"})
}

func withContainer(pod *corev1.Pod, nameSuffix string) *corev1.Pod {
	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{Name: pod.Name + nameSuffix})
	return pod
}
