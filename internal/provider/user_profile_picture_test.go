//go:build acc
// +build acc

package provider_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// writeTestPNG writes a small solid-colour PNG. Pocket ID decodes uploaded
// profile pictures, so the fixtures have to be real images.
func writeTestPNG(t *testing.T, p string, c color.Color) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encoding png: %v", err)
	}
	if err := os.WriteFile(p, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("writing png: %v", err)
	}
}

// getUserProfilePicture returns the profile picture Pocket ID serves for a
// user: the custom one if set, otherwise the default.
func getUserProfilePicture(id string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/users/%s/profile-picture.png", os.Getenv("POCKETID_BASE_URL"), id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// uploadUserProfilePicture uploads a profile picture directly through the
// API, the way a user would in the Pocket ID UI.
func uploadUserProfilePicture(id, p string) error {
	c, err := testClient()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return c.UpdateUserProfilePicture(id, filepath.Base(p), data)
}

// checkNoUserNamed asserts no user with the given username exists.
func checkNoUserNamed(username string) error {
	c, err := testClient()
	if err != nil {
		return err
	}

	users, err := c.ListUsers()
	if err != nil {
		return err
	}

	for _, u := range users.Data {
		if u.Username == username {
			return fmt.Errorf("user %q (%s) was left behind", username, u.ID)
		}
	}
	return nil
}
