/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cmd

// volumeInfoFormat is the default Go template format for emitting a
// csi.VolumeInfo
const volumeInfoFormat = `{{printf "%q\t%d" .VolumeId .CapacityBytes}}` +
	`{{if .VolumeContext}}{{"\t"}}` +
	`{{range $k, $v := .VolumeContext}}{{printf "%q=%q\t" $k $v}}{{end}}` +
	`{{end}}{{"\n"}}`

// volumeInfoFormat is the default Go template format for emitting a
// csi.SnapshotInfo
const snapshotInfoFormat = `{{printf "%q\t%d\t%s\t%s\t%t\n" ` +
	`.SnapshotId .SizeBytes .SourceVolumeId .CreationTime .ReadyToUse}}`

// listVolumesFormat is the default Go template format for emitting a
// ListVolumesResponse
const listVolumesFormat = `{{range $k, $v := .Entries}}` +
	`{{with $v.Volume}}` + volumeInfoFormat + `{{end}}` +
	`{{end}}` + // {{range $v .Entries}}
	`{{if .NextToken}}{{printf "token=%q\n" .NextToken}}{{end}}`

// listSnapshotsFormat is the default Go template format for emitting a
// ListSnapshotsResponse
const listSnapshotsFormat = `{{range $k, $s := .Entries}}` +
	`{{with $s.Snapshot}}` + snapshotInfoFormat + `{{end}}` +
	`{{end}}` + // {{range $s .Entries}}
	`{{if .NextToken}}{{printf "token=%q\n" .NextToken}}{{end}}`

// supportedVersionsFormat is the default Go template for emitting a
// csi.GetSupportedVersionsResponse
const supportedVersionsFormat = `{{range $v := .SupportedVersions}}` +
	`{{printf "%d.%d.%d\n" $v.Major $v.Minor $v.Patch}}{{end}}`

// pluginInfoFormat is the default Go template for emitting a
// csi.GetPluginInfoResponse
const pluginInfoFormat = `{{printf "%q\t%q" .Name .VendorVersion}}` +
	`{{range $k, $v := .Manifest}}{{printf "\t%q=%q" $k $v}}{{end}}` +
	`{{"\n"}}`

// pluginCapsFormat is the default Go template for emitting a
// csi.GetPluginCapabilities
const pluginCapsFormat = `{{range $v := .Capabilities}}` +
	`{{with $t := .Type}}` +
	`{{if isa $t "*csi.PluginCapability_Service_"}}{{if $t.Service}}` +
	`{{printf "%s\n" $t.Service.Type}}` +
	`{{end}}{{end}}` +
	`{{if isa $t "*csi.PluginCapability_VolumeExpansion_"}}{{if $t.VolumeExpansion}}` +
	`{{printf "%s\n" $t.VolumeExpansion.Type}}` +
	`{{end}}{{end}}` +
	`{{end}}` +
	`{{end}}`

// probeFormat is the default Go template for emitting a
// csi.Probe
const probeFormat = `{{printf "%t\n" .Ready.Value}}`

// statsFormat is the default Go template for emitting a
// csi.NodeGetVolumeStats
const statsFormat = `{{printf "%s\t%s\t" .Name .Path}}` +
	`{{range .Resp.Usage}}` +
	`{{printf "%d\t%d\t%d\t%s\n" .Available .Total .Used .Unit}}` +
	`{{end}}`

const nodeInfoFormat = `{{printf "%s\t%d\t%#v\n" .NodeId .MaxVolumesPerNode .AccessibleTopology}}`
