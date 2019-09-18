package simple

import (
	"context"
	"github.com/stretchr/testify/assert"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/utils"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"io"
	"sort"
	"testing"
)

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

func TestLocalVolumeServer_CreateVolume(t *testing.T) {
	t.Run("should_success_when_new_volume_creation", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())
			createVolumesByName(t, client, ctx, []string{"dummy-volume"})
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
			assert.Error(t, err, "volume name is duplicated")
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
			assert.Error(t, err, "volume not found")
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
			assert.Equal(t, ids[0], res.GetId())
			assert.Equal(t, "foo", res.GetInfo().GetName())
		})
	})
	t.Run("should_fail_when_inspect_not_existing_volume_by_name", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			res, err := client.InspectVolume(ctx, &elton_v2.InspectVolumeRequest{
				Name: "not-found",
			})
			assert.EqualError(t, err, "not found volume")
			assert.Nil(t, res)
		})
	})
	t.Run("should_fail_when_inspect_not_existing_volume_by_name", func(t *testing.T) {
		utils.WithTestServer(&Server{}, func(ctx context.Context, dial func() *grpc.ClientConn) {
			client := elton_v2.NewVolumeServiceClient(dial())

			res, err := client.InspectVolume(ctx, &elton_v2.InspectVolumeRequest{
				Id: &elton_v2.VolumeID{Id: "not-found"},
			})
			assert.EqualError(t, err, "not found volume")
			assert.Nil(t, res)
		})
	})
}
