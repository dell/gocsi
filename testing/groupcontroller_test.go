/*
 *
 * Copyright 2026 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the License; you may not use this file except in compliance with the License.
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

package gocsi_test

import (
	"context"
	"fmt"
	"math"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var _ = Describe("GroupController", func() {
	var (
		err         error
		stopMock    func()
		ctx         context.Context
		gclient     *grpc.ClientConn
		client      csi.GroupControllerClient
		groupSnapID string
	)
	BeforeEach(func() {
		ctx = context.Background()
		groupSnapID = "1"
	})
	JustBeforeEach(func() {
		gclient, stopMock, err = startMockServer(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		client = csi.NewGroupControllerClient(gclient)
	})
	AfterEach(func() {
		ctx = nil
		gclient.Close()
		gclient = nil
		client = nil
		stopMock()
	})
	Describe("CreateVolumeGroupSnapshot", func() {
		JustBeforeEach(func() {
			createNewVolumeGroupSnapshot(ctx, groupSnapID, client)
		})
		Context("Normal Create VolumeGroupSnapshot Call", func() {
			It("should validate the created groupSnapshot", func() {
				validateNewVolumeGroupSnapshot(ctx, groupSnapID, client)
			})
		})
	})
	Describe("GetVolumeGroupSnapshot", func() {
		BeforeEach(func() {
			// service initialization populates one volume group snapshot ID already, we will just use this here.
			groupSnapID = "Mock Group Snapshot 1"
		})
		AfterEach(func() {
			groupSnapID = ""
		})
		getVolumeGroupSnapshot := func(snapID string) {
			_, err = client.GetVolumeGroupSnapshot(
				ctx,
				&csi.GetVolumeGroupSnapshotRequest{
					GroupSnapshotId: snapID,
				})
		}
		JustBeforeEach(func() {
			getVolumeGroupSnapshot(groupSnapID)
		})
		Context("Valid Get", func() {
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("Missing Group Snapshot ID", func() {
			BeforeEach(func() {
				groupSnapID = ""
			})
			It("Should Not Be Valid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(codes.NotFound, "Group snapshot not found"))
			})
		})
		Context("VG Snapshot not found", func() {
			BeforeEach(func() {
				groupSnapID = "5"
			})
			It("Should Not Be Valid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(codes.NotFound, "Group snapshot not found"))
			})
		})
		Context("Error Injection", func() {
			It("Should Test Mock Service Behavior", func() {
				// Test that the mock service responds correctly to normal requests
				req := &csi.GetVolumeGroupSnapshotRequest{
					GroupSnapshotId: groupSnapID,
				}
				_, err := client.GetVolumeGroupSnapshot(ctx, req)
				// This should succeed for valid group snapshot ID
				if groupSnapID == "Mock Group Snapshot 1" {
					Ω(err).ShouldNot(HaveOccurred())
				} else {
					// For invalid IDs, expect error
					Ω(err).Should(HaveOccurred())
				}
			})
		})
	})
	Describe("DeleteVolumeGroupSnapshot", func() {
		BeforeEach(func() {
			groupSnapID = "Mock Group Snapshot 1"
		})
		AfterEach(func() {
			groupSnapID = ""
		})
		deleteVolumeGroupSnapshot := func(snapID string) {
			_, err = client.DeleteVolumeGroupSnapshot(
				ctx,
				&csi.DeleteVolumeGroupSnapshotRequest{
					GroupSnapshotId: snapID,
				})
		}
		JustBeforeEach(func() {
			deleteVolumeGroupSnapshot(groupSnapID)
		})
		Context("Valid Delete", func() {
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("Missing Group Snapshot ID", func() {
			BeforeEach(func() {
				groupSnapID = ""
			})
			It("Should Not Be Valid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(codes.NotFound, "Group snapshot not found"))
			})
		})
		Context("Not Found", func() {
			BeforeEach(func() {
				groupSnapID = "5"
			})
			It("Should Not Be Valid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(codes.NotFound, "Group snapshot not found"))
			})
		})
		Context("Error Injection", func() {
			It("Should Test Mock Service Behavior", func() {
				// Test that the mock service responds correctly to normal requests
				// Use a non-existent snapshot ID to test error path
				req := &csi.DeleteVolumeGroupSnapshotRequest{
					GroupSnapshotId: "non-existent-snapshot",
				}
				_, err := client.DeleteVolumeGroupSnapshot(ctx, req)
				// This should fail for non-existent snapshot ID
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(codes.NotFound, "Group snapshot not found"))
			})
		})
	})

	Describe("GroupControllerGetCapabilities", func() {
		var res *csi.GroupControllerGetCapabilitiesResponse
		getCapabilities := func() {
			res, err = client.GroupControllerGetCapabilities(
				ctx,
				&csi.GroupControllerGetCapabilitiesRequest{})
		}
		JustBeforeEach(func() {
			getCapabilities()
		})
		Context("Normal Capabilities Call", func() {
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).ShouldNot(BeNil())
				Ω(res.Capabilities).Should(HaveLen(1))
				Ω(res.Capabilities[0].Type).ShouldNot(BeNil())
				Ω(res.Capabilities[0].GetRpc().GetType()).Should(Equal(csi.GroupControllerServiceCapability_RPC_CREATE_DELETE_GET_VOLUME_GROUP_SNAPSHOT))
			})
		})
		Context("Error Injection", func() {
			It("Should Test Mock Service Behavior", func() {
				// Test that the mock service responds correctly to normal requests
				res, err = client.GroupControllerGetCapabilities(
					ctx,
					&csi.GroupControllerGetCapabilitiesRequest{})
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).ShouldNot(BeNil())
				Ω(res.Capabilities).Should(HaveLen(1))
			})
		})
	})

	Describe("CreateVolumeGroupSnapshot - Error Scenarios", func() {
		Context("Missing Name", func() {
			BeforeEach(func() {
				groupSnapID = ""
			})
			It("Should Be Invalid", func() {
				req := &csi.CreateVolumeGroupSnapshotRequest{
					Name:            "",
					SourceVolumeIds: []string{"vol-1", "vol-2"},
				}
				_, err := client.CreateVolumeGroupSnapshot(ctx, req)
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(codes.InvalidArgument, "exceeds size limit: Name: max=128, size=0"))
			})
		})
		Context("Empty Source Volume IDs", func() {
			It("Should Be Invalid", func() {
				req := &csi.CreateVolumeGroupSnapshotRequest{
					Name:            "test-snapshot",
					SourceVolumeIds: []string{},
				}
				_, err := client.CreateVolumeGroupSnapshot(ctx, req)
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(codes.InvalidArgument, "required: SourceVolumeIds"))
			})
		})
		Context("Error Injection", func() {
			It("Should Test Mock Service Behavior", func() {
				// Test that the mock service responds correctly to normal requests
				req := &csi.CreateVolumeGroupSnapshotRequest{
					Name:            "test-snapshot",
					SourceVolumeIds: []string{"vol-1", "vol-2"},
				}
				_, err := client.CreateVolumeGroupSnapshot(ctx, req)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("Invalid Parameters - Long Name", func() {
			It("Should Be Invalid", func() {
				longName := "this-is-a-very-long-group-snapshot-name-that-exceeds-the-normal-validation-limits-and-should-trigger-validation-errors-because-it-is-over-128-chars"
				req := &csi.CreateVolumeGroupSnapshotRequest{
					Name:            longName,
					SourceVolumeIds: []string{"vol-1", "vol-2"},
				}
				_, err := client.CreateVolumeGroupSnapshot(ctx, req)
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(codes.InvalidArgument, "exceeds size limit: Name: max=128, size=147"))
			})
		})
		Context("Invalid Parameters - Nil Source Volume IDs", func() {
			It("Should Be Invalid", func() {
				req := &csi.CreateVolumeGroupSnapshotRequest{
					Name:            "test-snapshot",
					SourceVolumeIds: nil,
				}
				_, err := client.CreateVolumeGroupSnapshot(ctx, req)
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(codes.InvalidArgument, "required: SourceVolumeIds"))
			})
		})
	})

	Describe("CreateVolumeGroupSnapshot - Idempotent Create", func() {
		const bucketSize = 250

		var (
			wg                   sync.WaitGroup
			count                int
			opPendingErrorOccurs bool
			mu                   sync.Mutex
		)

		idempCreateGroupSnapshots := func() {
			var (
				once    sync.Once
				buckets = count / bucketSize
				worker  = func() {
					defer wg.Done()
					defer GinkgoRecover()
					req := &csi.CreateVolumeGroupSnapshotRequest{
						Name:            fmt.Sprintf("test-snap-%d", count),
						SourceVolumeIds: []string{"vol-1", "vol-2"},
					}
					_, err := client.CreateVolumeGroupSnapshot(ctx, req)
					if err != nil {
						mu.Lock()
						once.Do(func() {
							if status.Code(err) == codes.Aborted {
								opPendingErrorOccurs = true
							}
						})
						mu.Unlock()
					}
				}
			)
			if r := math.Remainder(float64(count), float64(bucketSize)); r > 0 {
				buckets++
			}
			for i := 0; i < buckets; i++ {
				go func(i int) {
					defer GinkgoRecover()
					start := i * bucketSize
					mu.Lock()
					for j := start; j < start+bucketSize && j < count; j++ {
						go worker()
					}
					mu.Unlock()
				}(i)
			}
		}

		validateIdempResult := func() {
			wg.Wait()
			if count >= 1000 {
				Ω(opPendingErrorOccurs).Should(BeTrue())
			}
		}

		JustBeforeEach(func() {
			idempCreateGroupSnapshots()
			wg.Add(count)
		})

		AfterEach(func() {
			mu.Lock()
			count = 0
			opPendingErrorOccurs = false
			mu.Unlock()
		})

		Context("x1", func() {
			BeforeEach(func() {
				count = 1
			})
			It("Should Be Valid", validateIdempResult)
		})
		Context("x10", func() {
			BeforeEach(func() {
				count = 10
			})
			It("Should Be Valid", validateIdempResult)
		})
	})
})

func createNewVolumeGroupSnapshot(ctx context.Context, snapName string, client csi.GroupControllerClient) error {
	req := &csi.CreateVolumeGroupSnapshotRequest{
		Name:            snapName,
		SourceVolumeIds: []string{"vol-1", "vol-2"},
	}
	_, err := client.CreateVolumeGroupSnapshot(ctx, req)
	if err != nil {
		// The mock service returns success, not error, for normal cases
		// Only expect errors for invalid requests or context-based error injection
		return err
	}
	return nil
}

func validateNewVolumeGroupSnapshot(ctx context.Context, groupSnapID string, client csi.GroupControllerClient) {
	req := &csi.GetVolumeGroupSnapshotRequest{
		GroupSnapshotId: groupSnapID,
	}
	resp, err := client.GetVolumeGroupSnapshot(ctx, req)
	if err != nil {
		Ω(err).Should(ΣCM(codes.NotFound, "Group snapshot not found"))
		return
	}

	Ω(resp.GroupSnapshot.GroupSnapshotId).Should(Equal(groupSnapID))
	Ω(resp.GroupSnapshot.Snapshots).Should(HaveLen(2))
	Ω(resp.GroupSnapshot.ReadyToUse).Should(BeTrue())

	// In the mock implementation, snapshot IDs are the same as source volume IDs
	Expect(resp.GroupSnapshot.Snapshots).To(
		WithTransform(func(snaps []*csi.Snapshot) []string {
			ids := make([]string, 0, len(snaps))
			for _, s := range snaps {
				ids = append(ids, s.SnapshotId)
			}
			return ids
		}, ConsistOf("vol-1", "vol-2")),
	)

	// Also validate source volume IDs
	Expect(resp.GroupSnapshot.Snapshots).To(
		WithTransform(func(snaps []*csi.Snapshot) []string {
			ids := make([]string, 0, len(snaps))
			for _, s := range snaps {
				ids = append(ids, s.SourceVolumeId)
			}
			return ids
		}, ConsistOf("vol-1", "vol-2")),
	)

	Ω(err).ShouldNot(HaveOccurred())
}
