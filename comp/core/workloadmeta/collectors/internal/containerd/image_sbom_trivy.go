// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build containerd && trivy

package containerd

import (
	"context"
	"fmt"

	"github.com/CycloneDX/cyclonedx-go"

	"github.com/DataDog/datadog-agent/comp/core/workloadmeta"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/sbom"
	"github.com/DataDog/datadog-agent/pkg/sbom/collectors/containerd"
	"github.com/DataDog/datadog-agent/pkg/sbom/scanner"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

func sbomCollectionIsEnabled() bool {
	return imageMetadataCollectionIsEnabled() && config.Datadog.GetBool("sbom.container_image.enabled")
}

func (c *collector) startSBOMCollection(ctx context.Context) error {
	if !sbomCollectionIsEnabled() {
		return nil
	}

	c.scanOptions = sbom.ScanOptionsFromConfig(config.Datadog, true)
	c.sbomScanner = scanner.GetGlobalScanner()
	if c.sbomScanner == nil {
		return fmt.Errorf("error retrieving global SBOM scanner")
	}

	filterParams := workloadmeta.FilterParams{
		Kinds:     []workloadmeta.Kind{workloadmeta.KindContainerImageMetadata},
		Source:    workloadmeta.SourceAll,
		EventType: workloadmeta.EventTypeSet,
	}
	imgEventsCh := c.store.Subscribe(
		"SBOM collector",
		workloadmeta.NormalPriority,
		workloadmeta.NewFilter(&filterParams),
	)
	resultChan := make(chan sbom.ScanResult, 2000)
	go func() {
		for {
			select {
			// We don't want to keep scanning if image channel is not empty but context is expired
			case <-ctx.Done():
				close(resultChan)
				return

			case eventBundle := <-imgEventsCh:
				close(eventBundle.Ch)

				for _, event := range eventBundle.Events {
					image := event.Entity.(*workloadmeta.ContainerImageMetadata)

					if image.SBOM.Status != workloadmeta.Pending {
						// BOM already stored. Can happen when the same image ID
						// is referenced with different names.
						log.Debugf("Image: %s/%s (id %s) SBOM already available", image.Namespace, image.Name, image.ID)
						continue
					}

					if err := c.extractSBOMWithTrivy(ctx, image, resultChan); err != nil {
						log.Warnf("Error extracting SBOM for image: namespace=%s name=%s, err: %s", image.Namespace, image.Name, err)
					}
				}
			}
		}
	}()

	go func() {
		for result := range resultChan {
			if result.ImgMeta == nil {
				log.Errorf("Scan result does not hold the image identifier. Error: %s", result.Error)
				continue
			}

			status := workloadmeta.Success
			reportedError := ""
			var report *cyclonedx.BOM
			if result.Error != nil {
				// TODO: add a retry mechanism for retryable errors
				log.Errorf("Failed to generate SBOM for containerd image: %s", result.Error)
				status = workloadmeta.Failed
				reportedError = result.Error.Error()
			} else {
				bom, err := result.Report.ToCycloneDX()
				if err != nil {
					log.Errorf("Failed to extract SBOM from report")
					status = workloadmeta.Failed
					reportedError = err.Error()
				}
				report = bom
			}

			sbom := &workloadmeta.SBOM{
				CycloneDXBOM:       report,
				GenerationTime:     result.CreatedAt,
				GenerationDuration: result.Duration,
				Status:             status,
				Error:              reportedError,
			}

			// Updating workloadmeta entities directly is not thread-safe, that's why we
			// generate an update event here instead.
			if err := c.handleImageCreateOrUpdate(ctx, result.ImgMeta.Namespace, result.ImgMeta.Name, sbom); err != nil {
				log.Warnf("Error extracting SBOM for image: namespace=%s name=%s, err: %s", result.ImgMeta.Namespace, result.ImgMeta.Name, err)
			}
		}
	}()

	return nil
}

func (c *collector) extractSBOMWithTrivy(_ context.Context, storedImage *workloadmeta.ContainerImageMetadata, resultChan chan<- sbom.ScanResult) error {
	containerdImage, err := c.containerdClient.Image(storedImage.Namespace, storedImage.Name)
	if err != nil {
		return err
	}

	scanRequest := &containerd.ScanRequest{
		Image:            containerdImage,
		ImageMeta:        storedImage,
		ContainerdClient: c.containerdClient,
		FromFilesystem:   config.Datadog.GetBool("sbom.container_image.use_mount"),
	}
	if err = c.sbomScanner.Scan(scanRequest, c.scanOptions, resultChan); err != nil {
		log.Errorf("Failed to trigger SBOM generation for containerd: %s", err)
		return err
	}

	return nil
}
