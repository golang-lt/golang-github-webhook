package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type hook struct {
	ID, Secret string
	Command    struct {
		Workdir, Exec string
	}
}

func (h *hook) authorized(req *http.Request, payload []byte) error {
	signature := strings.Split(req.Header.Get("X-Hub-Signature"), "=")
	if len(signature) != 2 {
		return fmt.Errorf("signature header is not provided, or is not valid: %s", signature)
	}
	var hasher func() hash.Hash
	switch signature[0] {
	case "sha1":
		hasher = sha1.New
	default:
		return fmt.Errorf("unidentified hash method provided: %s", signature[0])
	}

	mac := hmac.New(hasher, []byte(h.Secret))
	_, err := mac.Write(payload)
	if err != nil {
		return fmt.Errorf("failed to write hmac payload: %s", err)
	}
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature[1]), []byte(expected)) {
		return fmt.Errorf("could not match signatures")
	}
	return nil
}

func (h *hook) run(p *payload) error {
	args := []string{
		p.Pusher.Name,
		p.Pusher.Email,
		p.Commit.ID,
		p.Commit.Message,
		p.Commit.Timestamp,
	}
	cmd := exec.Command(h.Command.Exec, args...)
	cmd.Dir = h.Command.Workdir

	info := fmt.Sprintf(`run hook "%s" workdir "%s" command: %s`, h.ID, cmd.Dir, h.Command.Exec)
	log.Println(info, strings.Join(args, " "))

	out, err := cmd.CombinedOutput()
	log.Println("command output:", string(out))
	return err
}
