package auth

import (
	"context"
	"github.com/fluxcd/go-git/v5/plumbing/transport/http"
	"github.com/fluxcd/go-git/v5/plumbing/transport/ssh"
	"github.com/kluctl/kluctl/v2/pkg/git/git-url"
	"github.com/kluctl/kluctl/v2/pkg/git/messages"
	ssh2 "golang.org/x/crypto/ssh"
	"strings"
)

type ListAuthProvider struct {
	MessageCallbacks messages.MessageCallbacks

	entries []AuthEntry
}

type AuthEntry struct {
	Host       string
	PathPrefix string
	Username   string
	Password   string

	SshKey     []byte
	KnownHosts []byte

	CABundle []byte
}

type KnownHostsWrapper struct {
	authMethod      ssh.AuthMethod
	hostKeyCallback ssh2.HostKeyCallback
}

func (w *KnownHostsWrapper) String() string {
	return w.authMethod.String()
}

func (w *KnownHostsWrapper) Name() string {
	return w.authMethod.Name()
}

func (w *KnownHostsWrapper) ClientConfig() (*ssh2.ClientConfig, error) {
	ccfg, err := w.authMethod.ClientConfig()
	if err != nil {
		return nil, err
	}
	ccfg.HostKeyCallback = w.hostKeyCallback
	return ccfg, nil
}

func (a *ListAuthProvider) AddEntry(e AuthEntry) {
	a.entries = append(a.entries, e)
}

func (a *ListAuthProvider) BuildAuth(ctx context.Context, gitUrl git_url.GitUrl) AuthMethodAndCA {
	a.MessageCallbacks.Trace("ListAuthProvider: BuildAuth for %s", gitUrl.String())
	a.MessageCallbacks.Trace("ListAuthProvider: path=%s, username=%s, scheme=%s", gitUrl.Path, gitUrl.User.Username(), gitUrl.Scheme)
	for _, e := range a.entries {
		a.MessageCallbacks.Trace("ListAuthProvider: try host=%s, pathPrefix=%s, username=%s", e.Host, e.PathPrefix, e.Username)

		if e.Host != "*" && e.Host != gitUrl.Hostname() {
			continue
		}
		urlPath := gitUrl.Path
		if strings.HasPrefix(urlPath, "/") {
			urlPath = urlPath[1:]
		}
		if !strings.HasPrefix(urlPath, e.PathPrefix) {
			continue
		}
		if e.Username == "" {
			continue
		}

		username := ""
		if gitUrl.User != nil {
			username = gitUrl.User.Username()
		}

		if username != "" && e.Username != "*" && username != e.Username {
			continue
		}

		if username == "" {
			username = e.Username
		}

		if username == "*" {
			// can't use "*" as username
			continue
		}

		if gitUrl.IsSsh() {
			if e.SshKey == nil {
				a.MessageCallbacks.Trace("ListAuthProvider: empty ssh key is not accepted")
				continue
			}
			a.MessageCallbacks.Trace("ListAuthProvider: using username+sshKey")
			pk, err := ssh.NewPublicKeys(username, e.SshKey, "")
			if err != nil {
				a.MessageCallbacks.Trace("ListAuthProvider: failed to parse private key: %v", err)
			} else {
				return AuthMethodAndCA{
					AuthMethod: &KnownHostsWrapper{
						authMethod:      pk,
						hostKeyCallback: buildVerifyHostCallback(a.MessageCallbacks, e.KnownHosts),
					},
					Hash: func() ([]byte, error) {
						return buildHash(pk.Signer)
					},
				}
			}
		} else {
			if e.Password == "" {
				a.MessageCallbacks.Trace("ListAuthProvider: empty password is not accepted")
				continue
			}
			a.MessageCallbacks.Trace("ListAuthProvider: using username+password")
			return AuthMethodAndCA{
				AuthMethod: &http.BasicAuth{
					Username: username,
					Password: e.Password,
				},
				CABundle: e.CABundle,
			}
		}
	}
	return AuthMethodAndCA{}
}
