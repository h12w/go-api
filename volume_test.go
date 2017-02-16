package storageos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/storageos/go-api/types"
)

func TestVolumeList(t *testing.T) {
	volumesData := `[{
    "id": "6b7afe82-565f-a627-4696-22457da5da9c",
    "master": {
        "controller": "",
        "id": "",
        "inode": 0,
        "status": "",
        "health": "",
        "created_at": "0001-01-01T00:00:00Z"
    },
    "replicas": null,
    "created_by": "storageos",
    "datacentre": "",
    "tenant": "",
    "name": "test02",
    "status": "pending",
    "status_message": "",
    "health": "",
    "pool": "213498fb-ead9-2a48-92e6-4dac2020f2ed",
    "description": "",
    "size": 10,
    "inode": 0,
    "volume_groups": [],
    "tags": ["filesystem"],
    "mounted": false,
    "no_of_mounts": 0,
    "mounted_by": "",
    "mounted_at": "0001-01-01T00:00:00Z",
    "created_at": "0001-01-01T00:00:00Z"
}, {
    "id": "ef897b9f-0b47-08ee-b669-0a2057df981c",
    "master": {
        "controller": "b3eb8d63-4f1b-9ef5-a504-7d02d604feb4",
        "id": "55fb06cb-263d-08bf-584e-e5b889166f3b",
        "inode": 41560,
        "status": "active",
        "health": "",
        "created_at": "2017-01-25T02:17:05.507557244Z"
    },
    "replicas": null,
    "created_by": "storageos",
    "datacentre": "",
    "tenant": "",
    "name": "test01",
    "status": "active",
    "status_message": "",
    "health": "",
    "pool": "213498fb-ead9-2a48-92e6-4dac2020f2ed",
    "description": "",
    "size": 10,
    "inode": 41397,
    "volume_groups": null,
    "tags": ["filesystem", "compression"],
    "mounted": false,
    "no_of_mounts": 0,
    "mounted_by": "",
    "mounted_at": "0001-01-01T00:00:00Z",
    "created_at": "0001-01-01T00:00:00Z"
}]`

	var expected []types.Volume
	if err := json.Unmarshal([]byte(volumesData), &expected); err != nil {
		t.Fatal(err)
	}

	client := newTestClient(&FakeRoundTripper{message: volumesData, status: http.StatusOK})
	volumes, err := client.VolumeList(types.ListOptions{})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(volumes, expected) {
		t.Errorf("Volumes: Wrong return value. Want %#v. Got %#v.", expected, volumes)
	}
}

func TestVolumeCreate(t *testing.T) {
	message := "\"ef897b9f-0b47-08ee-b669-0a2057df981c\""
	fakeRT := &FakeRoundTripper{message: message, status: http.StatusOK}
	client := newTestClient(fakeRT)
	id, err := client.VolumeCreate(
		types.VolumeCreateOptions{
			Name:        "unit01",
			Description: "Unit test volume",
			Pool:        "default",
			Labels: map[string]string{
				"foo": "bar",
			},
			Context: context.Background(),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 36 {
		t.Errorf("VolumeCreate: Wrong return value. Wanted 34 character UUID. Got %d. (%s)", len(id), id)
	}
	req := fakeRT.requests[0]
	expectedMethod := "POST"
	if req.Method != expectedMethod {
		t.Errorf("VolumeCreate(): Wrong HTTP method. Want %s. Got %s.", expectedMethod, req.Method)
	}
	u, _ := url.Parse(client.getURL(VolumeAPIPrefix))
	if req.URL.Path != u.Path {
		t.Errorf("VolumeCreate(): Wrong request path. Want %q. Got %q.", u.Path, req.URL.Path)
	}
}

func TestVolume(t *testing.T) {
	body := `{
		"Name": "unit01",
		"Description": "Unit test volume",
		"Pool": "default",
		"Size": 5
	}`
	var expected types.Volume
	if err := json.Unmarshal([]byte(body), &expected); err != nil {
		t.Fatal(err)
	}
	fakeRT := &FakeRoundTripper{message: body, status: http.StatusOK}
	client := newTestClient(fakeRT)
	name := "tardis"
	volume, err := client.Volume(name)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(volume, &expected) {
		t.Errorf("Volume: Wrong return value. Want %#v. Got %#v.", expected, volume)
	}
	req := fakeRT.requests[0]
	expectedMethod := "GET"
	if req.Method != expectedMethod {
		t.Errorf("InspectVolume(%q): Wrong HTTP method. Want %s. Got %s.", name, expectedMethod, req.Method)
	}
	u, _ := url.Parse(client.getURL(VolumeAPIPrefix + "/" + name))
	if req.URL.Path != u.Path {
		t.Errorf("VolumeCreate(%q): Wrong request path. Want %q. Got %q.", name, u.Path, req.URL.Path)
	}
}

func TestVolumeDelete(t *testing.T) {
	name := "test"
	fakeRT := &FakeRoundTripper{message: "", status: http.StatusNoContent}
	client := newTestClient(fakeRT)
	if err := client.VolumeDelete(name); err != nil {
		t.Fatal(err)
	}
	req := fakeRT.requests[0]
	expectedMethod := "DELETE"
	if req.Method != expectedMethod {
		t.Errorf("VolumeDelete(%q): Wrong HTTP method. Want %s. Got %s.", name, expectedMethod, req.Method)
	}
	u, _ := url.Parse(client.getURL(VolumeAPIPrefix + "/" + name))
	if req.URL.Path != u.Path {
		t.Errorf("VolumeDelete(%q): Wrong request path. Want %q. Got %q.", name, u.Path, req.URL.Path)
	}
}

func TestVolumeDeleteNotFound(t *testing.T) {
	client := newTestClient(&FakeRoundTripper{message: "no such volume", status: http.StatusNotFound})
	if err := client.VolumeDelete("test:"); err != ErrNoSuchVolume {
		t.Errorf("VolumeDelete: wrong error. Want %#v. Got %#v.", ErrNoSuchVolume, err)
	}
}

func TestVolumeDeleteInUse(t *testing.T) {
	client := newTestClient(&FakeRoundTripper{message: "volume in use and cannot be removed", status: http.StatusConflict})
	if err := client.VolumeDelete("test:"); err != ErrVolumeInUse {
		t.Errorf("VolumeDelete: wrong error. Want %#v. Got %#v.", ErrVolumeInUse, err)
	}
}
