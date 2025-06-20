package bulk_clone

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitURLBuilder(t *testing.T) {
	t.Run("HTTPS URL construction", func(t *testing.T) {
		builder := NewGitURLBuilder("https", "github.com")
		url := builder.BuildURL("myorg", "myrepo")
		assert.Equal(t, "https://github.com/myorg/myrepo.git", url)
	})

	t.Run("SSH URL construction", func(t *testing.T) {
		builder := NewGitURLBuilder("ssh", "github.com")
		url := builder.BuildURL("myorg", "myrepo")
		assert.Equal(t, "git@github.com:myorg/myrepo.git", url)
	})

	t.Run("HTTP URL construction", func(t *testing.T) {
		builder := NewGitURLBuilder("http", "gitlab.example.com")
		url := builder.BuildURL("mygroup", "myproject")
		assert.Equal(t, "http://gitlab.example.com/mygroup/myproject.git", url)
	})

	t.Run("default to HTTPS for unknown protocol", func(t *testing.T) {
		builder := NewGitURLBuilder("unknown", "github.com")
		url := builder.BuildURL("myorg", "myrepo")
		assert.Equal(t, "https://github.com/myorg/myrepo.git", url)
	})
}

func TestSSHHostAlias(t *testing.T) {
	t.Run("GitHub SSH host alias", func(t *testing.T) {
		builder := NewGitURLBuilder("ssh", "github.com")
		alias := builder.BuildSSHHostAlias("mycompany")
		assert.Equal(t, "github-mycompany", alias)
	})

	t.Run("GitLab SSH host alias", func(t *testing.T) {
		builder := NewGitURLBuilder("ssh", "gitlab.com")
		alias := builder.BuildSSHHostAlias("mygroup")
		assert.Equal(t, "gitlab-mygroup", alias)
	})

	t.Run("Gitea SSH host alias", func(t *testing.T) {
		builder := NewGitURLBuilder("ssh", "gitea.com")
		alias := builder.BuildSSHHostAlias("myorg")
		assert.Equal(t, "gitea-myorg", alias)
	})

	t.Run("custom hostname returns as-is", func(t *testing.T) {
		builder := NewGitURLBuilder("ssh", "git.example.com")
		alias := builder.BuildSSHHostAlias("myorg")
		assert.Equal(t, "git.example.com", alias)
	})

	t.Run("non-SSH protocol returns original hostname", func(t *testing.T) {
		builder := NewGitURLBuilder("https", "github.com")
		alias := builder.BuildSSHHostAlias("myorg")
		assert.Equal(t, "github.com", alias)
	})
}

func TestBuildURLWithHostAlias(t *testing.T) {
	t.Run("SSH with host alias", func(t *testing.T) {
		builder := NewGitURLBuilder("ssh", "github.com")
		url := builder.BuildURLWithHostAlias("mycompany", "myrepo")
		assert.Equal(t, "git@github-mycompany:mycompany/myrepo.git", url)
	})

	t.Run("HTTPS without host alias", func(t *testing.T) {
		builder := NewGitURLBuilder("https", "github.com")
		url := builder.BuildURLWithHostAlias("mycompany", "myrepo")
		assert.Equal(t, "https://github.com/mycompany/myrepo.git", url)
	})

	t.Run("SSH with custom hostname", func(t *testing.T) {
		builder := NewGitURLBuilder("ssh", "git.example.com")
		url := builder.BuildURLWithHostAlias("myorg", "myrepo")
		assert.Equal(t, "git@git.example.com:myorg/myrepo.git", url)
	})
}

func TestGetDefaultHostname(t *testing.T) {
	tests := []struct {
		provider string
		expected string
	}{
		{"github", "github.com"},
		{"gitlab", "gitlab.com"},
		{"gitea", "gitea.com"},
		{"gogs", "gogs.com"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			hostname := GetDefaultHostname(tt.provider)
			assert.Equal(t, tt.expected, hostname)
		})
	}
}

func TestBuildURLForProvider(t *testing.T) {
	t.Run("GitHub HTTPS", func(t *testing.T) {
		url := BuildURLForProvider("github", "https", "myorg", "myrepo")
		assert.Equal(t, "https://github.com/myorg/myrepo.git", url)
	})

	t.Run("GitHub SSH", func(t *testing.T) {
		url := BuildURLForProvider("github", "ssh", "myorg", "myrepo")
		assert.Equal(t, "git@github.com:myorg/myrepo.git", url)
	})

	t.Run("GitLab HTTPS", func(t *testing.T) {
		url := BuildURLForProvider("gitlab", "https", "mygroup", "myproject")
		assert.Equal(t, "https://gitlab.com/mygroup/myproject.git", url)
	})

	t.Run("GitLab SSH", func(t *testing.T) {
		url := BuildURLForProvider("gitlab", "ssh", "mygroup", "myproject")
		assert.Equal(t, "git@gitlab.com:mygroup/myproject.git", url)
	})
}

func TestBuildURLWithHostAliasForProvider(t *testing.T) {
	t.Run("GitHub SSH with host alias", func(t *testing.T) {
		url := BuildURLWithHostAliasForProvider("github", "ssh", "mycompany", "myrepo")
		assert.Equal(t, "git@github-mycompany:mycompany/myrepo.git", url)
	})

	t.Run("GitHub HTTPS without host alias", func(t *testing.T) {
		url := BuildURLWithHostAliasForProvider("github", "https", "mycompany", "myrepo")
		assert.Equal(t, "https://github.com/mycompany/myrepo.git", url)
	})

	t.Run("GitLab SSH with host alias", func(t *testing.T) {
		url := BuildURLWithHostAliasForProvider("gitlab", "ssh", "mygroup", "myproject")
		assert.Equal(t, "git@gitlab-mygroup:mygroup/myproject.git", url)
	})

	t.Run("Gitea SSH with host alias", func(t *testing.T) {
		url := BuildURLWithHostAliasForProvider("gitea", "ssh", "myorg", "myrepo")
		assert.Equal(t, "git@gitea-myorg:myorg/myrepo.git", url)
	})

	t.Run("Gogs SSH with host alias", func(t *testing.T) {
		url := BuildURLWithHostAliasForProvider("gogs", "ssh", "myorg", "myrepo")
		assert.Equal(t, "git@gogs-myorg:myorg/myrepo.git", url)
	})
}
