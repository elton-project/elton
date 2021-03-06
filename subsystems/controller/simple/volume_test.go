package simple

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"math/rand"
	"sort"
	"testing"
)

func createVolume(t *testing.T, dial func() *grpc.ClientConn, ctx context.Context) *elton_v2.VolumeID {
	client := elton_v2.NewVolumeServiceClient(dial())
	res, err := client.CreateVolume(ctx, &elton_v2.CreateVolumeRequest{
		Info: &elton_v2.VolumeInfo{
			Name: "volume-" + fmt.Sprint(rand.Int63()),
		},
	})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("new VolumeID: %s", res.GetId().GetId())
	return res.Id
}
func createVolumesByName(t *testing.T, client elton_v2.VolumeServiceClient, ctx context.Context, names []string) ([]*elton_v2.VolumeID, error) {
	var ids []*elton_v2.VolumeID
	for _, name := range names {
		res, err := client.CreateVolume(ctx, &elton_v2.CreateVolumeRequest{
			Info: &elton_v2.VolumeInfo{Name: name},
		})
		if !assert.NoError(t, err) {
			return nil, err
		}
		t.Logf("new VolumeID: %s", res.GetId().GetId())
		if !assert.NotEmpty(t, res.GetId().GetId()) {
			return nil, xerrors.New("emtpy VolumeID")
		}
		ids = append(ids, res.GetId())
	}
	return ids, nil
}

func listVolumes(t *testing.T, client elton_v2.VolumeServiceClient, ctx context.Context) ([]string, error) {
	volumes := map[string]bool{}
	stream, err := client.ListVolumes(ctx, &elton_v2.ListVolumesRequest{})
	if !assert.NoError(t, err) {
		return nil, xerrors.Errorf(": %w", err)
	}
	for {
		volume, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if !assert.NoError(t, err) {
			return nil, err
		}

		name := volume.GetInfo().GetName()
		assert.NotEmpty(t, name)
		volumes[name] = true
	}

	var names []string
	for name := range volumes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}
func createCommits(
	t *testing.T,
	dial func() *grpc.ClientConn,
	ctx context.Context,
	volumeName string,
	commits []*elton_v2.CommitRequest,
) (*elton_v2.VolumeID, []*elton_v2.CommitID) {
	vc := elton_v2.NewVolumeServiceClient(dial())
	vres, err := vc.CreateVolume(ctx, &elton_v2.CreateVolumeRequest{
		Info: &elton_v2.VolumeInfo{
			Name: volumeName,
		},
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	volumeID := vres.GetId()
	assert.NotNil(t, volumeID)

	var ids []*elton_v2.CommitID
	cc := elton_v2.NewCommitServiceClient(dial())
	for _, commit := range commits {
		// Set VolumeID.
		commit.Id = volumeID
		// Set parent CommitID.
		if len(ids) > 0 {
			commit.Info.LeftParentID = ids[len(ids)-1]
		} else {
			lres, err := cc.GetLastCommit(ctx, &elton_v2.GetLastCommitRequest{
				VolumeId: volumeID,
			})
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			commit.Info.LeftParentID = lres.GetId()
		}
		assert.NotNil(t, commit.Info.LeftParentID)

		res, err := cc.Commit(ctx, commit)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		ids = append(ids, res.Id)
	}
	return volumeID, ids
}
func createEmptyTree() *elton_v2.Tree {
	return &elton_v2.Tree{
		RootIno: 1,
		Inodes: map[uint64]*elton_v2.File{
			1: {FileType: elton_v2.FileType_Directory},
		},
	}
}

func TestLocalVolumeServer_CreateVolume(t *testing.T) {
	t.Run("should_success_when_new_volume_creation", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())
			res, err := createVolumesByName(t, client, ctx, []string{"dummy-volume"})
			assert.NoError(t, err)
			if !assert.Len(t, res, 1) {
				return
			}
			assert.NotNil(t, res[0])
		})
	})
	t.Run("should_fail_when_creating_with_duplicated_name", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			// Step1: Create a volume.
			if _, err := createVolumesByName(t, client, ctx, []string{"foo"}); err != nil {
				return
			}

			// Step2: Create a volume with same name.
			res, err := client.CreateVolume(ctx, &elton_v2.CreateVolumeRequest{
				Info: &elton_v2.VolumeInfo{Name: "foo"},
			})
			assert.Equal(t, codes.AlreadyExists, status.Code(err))
			assert.Equal(t, "duplicate volume name: name=foo", status.Convert(err).Message())
			assert.Nil(t, res)
		})
	})
}

func TestLocalVolumeServer_DeleteVolume(t *testing.T) {
	t.Run("should_success_whne_delete_an existing_volume", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			// Step1: Create a volume.
			ids, err := createVolumesByName(t, client, ctx, []string{"foo"})
			if err != nil {
				return
			}

			// Step2: Delete an existing volume.
			_, err = client.DeleteVolume(ctx, &elton_v2.DeleteVolumeRequest{
				Id: ids[0],
			})
			assert.NoError(t, err)
		})
	})
	t.Run("should_fail_when_delete_not_existing_volume", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			_, err := client.DeleteVolume(ctx, &elton_v2.DeleteVolumeRequest{
				Id: &elton_v2.VolumeID{Id: "invalid-id"},
			})
			assert.Equal(t, codes.NotFound, status.Code(err))
			assert.Contains(t, status.Convert(err).Message(), "not found volume: ")
		})
	})
}

func TestLocalVolumeServer_ListVolumes(t *testing.T) {
	//t.Run("should_always_success_the_ListVolumes()", func(t *testing.T) {
	t.Run("should_success_when_list_the_emtpy_volume_list", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())
			stream, err := client.ListVolumes(ctx, &elton_v2.ListVolumesRequest{})
			assert.NoError(t, err)

			_, err = stream.Recv()
			assert.EqualError(t, err, io.EOF.Error())
		})
	})

	t.Run("should_success_volume_listing_(without_limit)", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())
			expectedVolumes := []string{
				"volume-1",
				"volume-2",
				"volume-3",
				"volume-4",
				"volume-5",
			}
			if _, err := createVolumesByName(t, client, ctx, expectedVolumes); err != nil {
				return
			}

			remoteVolumes, err := listVolumes(t, client, ctx)
			if err != nil {
				return
			}

			assert.Equal(t, expectedVolumes, remoteVolumes)
		})
	})
	t.Run("limiting_feature_should_work", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			expectedVolumes := []string{
				"volume-1",
				"volume-2",
				"volume-3",
				"volume-4",
				"volume-5",
			}

			if _, err := createVolumesByName(t, client, ctx, expectedVolumes); err != nil {
				return
			}

			remoteVolumes, err := client.ListVolumes(ctx, &elton_v2.ListVolumesRequest{
				Limit: 3,
			})
			if err != nil {
				return
			}

			count := 0
			for {
				vol, err := remoteVolumes.Recv()
				if err == io.EOF {
					break
				}
				if !assert.NoError(t, err) {
					return
				}
				count++
				name := vol.GetInfo().GetName()
				assert.Contains(t, expectedVolumes, name)
			}
			assert.Equal(t, 3, count)
		})
	})
	t.Run("should_fail_when_next_parameter_is_specified", func(t *testing.T) {
		// NOTE: local volume server is not supported of pagination feature.
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			stream, err := client.ListVolumes(ctx, &elton_v2.ListVolumesRequest{
				Next: "aaa",
			})
			assert.NoError(t, err)
			assert.NotNil(t, stream)

			res, err := stream.Recv()
			assert.Equal(t, status.Convert(err).Message(), "next parameter is not supported")
			assert.Equal(t, codes.FailedPrecondition, status.Code(err))
			assert.Nil(t, res)
		})
	})
}

func TestLocalVolumeServer_InspectVolume(t *testing.T) {
	t.Run("should_success_when_inspect_existing_volume_by_name", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			ids, err := createVolumesByName(t, client, ctx, []string{"foo"})
			if err != nil {
				return
			}

			res, err := client.InspectVolume(ctx, &elton_v2.InspectVolumeRequest{
				Name: "foo",
			})
			assert.NoError(t, err)
			assert.Equal(t, ids[0], res.GetId())
			assert.Equal(t, "foo", res.GetInfo().GetName())
		})
	})
	t.Run("should_success_when_inspect_existing_volume_by_id", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			ids, err := createVolumesByName(t, client, ctx, []string{"foo"})
			if err != nil {
				return
			}

			res, err := client.InspectVolume(ctx, &elton_v2.InspectVolumeRequest{
				Id: ids[0],
			})
			assert.NoError(t, err)
			assert.Equal(t, ids[0].String(), res.GetId().String())
			assert.Equal(t, "foo", res.GetInfo().GetName())
		})
	})
	t.Run("should_fail_when_inspect_not_existing_volume_by_name", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			res, err := client.InspectVolume(ctx, &elton_v2.InspectVolumeRequest{
				Name: "not-found",
			})

			assert.Equal(t, "not found volume: name=not-found", status.Convert(err).Message())
			assert.Equal(t, codes.NotFound, status.Code(err))
			assert.Nil(t, res)
		})
	})
	t.Run("should_fail_when_inspect_not_existing_volume_by_name", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			res, err := client.InspectVolume(ctx, &elton_v2.InspectVolumeRequest{
				Id: &elton_v2.VolumeID{Id: "not-found"},
			})
			assert.Contains(t, status.Convert(err).Message(), "not found volume: ")
			assert.Equal(t, codes.NotFound, status.Code(err))
			assert.Nil(t, res)
		})
	})
	t.Run("should_fail_when_id_and_name_is_not_specified", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			res, err := client.InspectVolume(ctx, &elton_v2.InspectVolumeRequest{
				Id:   &elton_v2.VolumeID{Id: "foo"},
				Name: "bar",
			})
			assert.Equal(t, status.Convert(err).Message(), "id and info is exclusive")
			assert.Equal(t, codes.FailedPrecondition, status.Code(err))
			assert.Nil(t, res)
		})
	})
	t.Run("should_fail_when_id_and_name_is_specified", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			res, err := client.InspectVolume(ctx, &elton_v2.InspectVolumeRequest{})
			assert.Equal(t, status.Convert(err).Message(), "id and info is exclusive")
			assert.Equal(t, codes.FailedPrecondition, status.Code(err))
			assert.Nil(t, res)
		})
	})
}

func TestLocalVolumeServer_GetLastCommit(t *testing.T) {
	t.Run("should_success_when_valid_volume_id", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			volume, commits := createCommits(t, dial, ctx, "test-volume", []*elton_v2.CommitRequest{
				{
					Info: &elton_v2.CommitInfo{
						CreatedAt: ptypes.TimestampNow(),
						Tree:      createEmptyTree(),
					},
				},
			})
			assert.NotNil(t, volume)
			assert.Len(t, commits, 1)

			client := elton_v2.NewCommitServiceClient(dial())
			res, err := client.GetLastCommit(ctx, &elton_v2.GetLastCommitRequest{
				VolumeId: volume,
			})
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, commits[0], res.GetId())
			assert.NotNil(t, res.GetInfo())
		})
	})
	t.Run("should_return_first_commit_id", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			vc := elton_v2.NewVolumeServiceClient(dial())
			vres, err := vc.CreateVolume(ctx, &elton_v2.CreateVolumeRequest{
				Info: &elton_v2.VolumeInfo{Name: "test-volume"},
			})
			assert.NoError(t, err)
			volume := vres.GetId()
			assert.NotNil(t, volume)

			client := elton_v2.NewCommitServiceClient(dial())
			res, err := client.GetLastCommit(ctx, &elton_v2.GetLastCommitRequest{
				VolumeId: volume,
			})
			assert.NoError(t, err)
			assert.NotNil(t, res)
		})
	})
	t.Run("should_fail_when_invalid_volume_id", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewCommitServiceClient(dial())
			res, err := client.GetLastCommit(ctx, &elton_v2.GetLastCommitRequest{})
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Nil(t, res)
		})
	})
}

func TestLocalVolumeServer_ListCommits(t *testing.T) {
	t.Run("should_fail_when_requesting_with_pagination", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewCommitServiceClient(dial())
			stream, err := client.ListCommits(ctx, &elton_v2.ListCommitsRequest{
				Next: "invalid",
			})
			assert.NoError(t, err)
			assert.NotNil(t, stream)
			res, err := stream.Recv()
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Nil(t, res)
		})
	})
	t.Run("should_success_when_volume_id_is_valid", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			vc := elton_v2.NewVolumeServiceClient(dial())
			vres, err := vc.CreateVolume(ctx, &elton_v2.CreateVolumeRequest{
				Info: &elton_v2.VolumeInfo{Name: "test-volume"},
			})
			assert.NoError(t, err)
			volume := vres.GetId()
			assert.NotNil(t, volume)

			client := elton_v2.NewCommitServiceClient(dial())
			stream, err := client.ListCommits(ctx, &elton_v2.ListCommitsRequest{
				Id: volume,
			})
			assert.NoError(t, err)
			assert.NotNil(t, stream)

			// Receive first commit.
			res, err := stream.Recv()
			assert.NoError(t, err)
			assert.NotNil(t, res)

			// Receive EOF.
			res, err = stream.Recv()
			assert.Equal(t, io.EOF, err)
			assert.Nil(t, res)
		})

	})
	t.Run("should_fail_when_volume_id_is_invalid", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewCommitServiceClient(dial())
			stream, err := client.ListCommits(ctx, &elton_v2.ListCommitsRequest{
				Id: &elton_v2.VolumeID{
					Id: "not-found",
				},
			})
			assert.NoError(t, err)
			assert.NotNil(t, stream)
			res, err := stream.Recv()
			assert.EqualError(t, err, io.EOF.Error())
			assert.Nil(t, res)
		})
	})
	t.Run("should_work_of_limit_args", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			volume, commits := createCommits(t, dial, ctx, "test-volume", []*elton_v2.CommitRequest{
				{
					Info: &elton_v2.CommitInfo{
						CreatedAt: ptypes.TimestampNow(),
						Tree:      createEmptyTree(),
					},
				}, {
					Info: &elton_v2.CommitInfo{
						CreatedAt: ptypes.TimestampNow(),
						Tree:      createEmptyTree(),
					},
				},
			})

			client := elton_v2.NewCommitServiceClient(dial())
			stream, err := client.ListCommits(ctx, &elton_v2.ListCommitsRequest{
				Id:    volume,
				Limit: 1,
			})
			if !assert.NoError(t, err) {
				return
			}

			latestCID := commits[len(commits)-1]
			res, err := stream.Recv()
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, latestCID, res.Id)

			res, err = stream.Recv()
			assert.EqualError(t, err, io.EOF.Error())
			assert.Nil(t, res)
		})
	})
}

func TestLocalVolumeServer_Commit(t *testing.T) {
	t.Run("should_success_when_creating_second_commit", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			volume, commits := createCommits(t, dial, ctx, "test-volume", []*elton_v2.CommitRequest{
				{
					Info: &elton_v2.CommitInfo{
						CreatedAt: ptypes.TimestampNow(),
						Tree:      createEmptyTree(),
					},
				}, {
					Info: &elton_v2.CommitInfo{
						CreatedAt: ptypes.TimestampNow(),
						Tree:      createEmptyTree(),
					},
				},
			})
			assert.NotEmpty(t, volume)
			assert.Len(t, commits, 2)
		})
	})
	t.Run("should_fail_when_parent_id_combination_is_invalid", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewCommitServiceClient(dial())
			res, err := client.Commit(ctx, &elton_v2.CommitRequest{
				Id: &elton_v2.VolumeID{Id: "test-volume"},
				Info: &elton_v2.CommitInfo{
					LeftParentID:  nil,
					RightParentID: &elton_v2.CommitID{Id: &elton_v2.VolumeID{Id: "foo"}, Number: 1},
					Tree:          createEmptyTree(),
				},
			})
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Equal(t, "left parent: id should not nil", status.Convert(err).Message())
			assert.Nil(t, res)
		})
	})
	t.Run("should_fail_when_non_existent_parent_id_is_specified", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			volume := createVolume(t, dial, ctx)
			client := elton_v2.NewCommitServiceClient(dial())
			res, err := client.Commit(ctx, &elton_v2.CommitRequest{
				Id: volume,
				Info: &elton_v2.CommitInfo{
					LeftParentID: &elton_v2.CommitID{Id: volume, Number: 1},
					Tree:         createEmptyTree(),
				},
			})
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Contains(t, status.Convert(err).Message(), "left parent: not found commit: ")
			assert.Nil(t, res)
		})
	})
}
