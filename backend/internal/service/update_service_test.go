//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type updateServiceCacheStub struct {
	data string
}

func (s *updateServiceCacheStub) GetUpdateInfo(context.Context) (string, error) {
	if s.data == "" {
		return "", errors.New("cache miss")
	}
	return s.data, nil
}

func (s *updateServiceCacheStub) SetUpdateInfo(_ context.Context, data string, _ time.Duration) error {
	s.data = data
	return nil
}

type updateServiceGitHubClientStub struct {
	release        *GitHubRelease
	recentReleases []*GitHubRelease
	recentErr      error
	latestRepo     string
	recentRepo     string
}

func (s *updateServiceGitHubClientStub) FetchLatestRelease(_ context.Context, repo string) (*GitHubRelease, error) {
	s.latestRepo = repo
	return s.release, nil
}

func (s *updateServiceGitHubClientStub) FetchRecentReleases(_ context.Context, repo string, _ int) ([]*GitHubRelease, error) {
	s.recentRepo = repo
	return s.recentReleases, s.recentErr
}

func (s *updateServiceGitHubClientStub) DownloadFile(context.Context, string, string, int64) error {
	panic("DownloadFile should not be called when no update is available")
}

func (s *updateServiceGitHubClientStub) FetchChecksumFile(context.Context, string) ([]byte, error) {
	panic("FetchChecksumFile should not be called when no update is available")
}

func TestUpdateServicePerformUpdateNoUpdateReturnsSentinel(t *testing.T) {
	svc := NewUpdateService(
		&updateServiceCacheStub{},
		&updateServiceGitHubClientStub{
			release: &GitHubRelease{
				TagName: "v0.1.132-1",
				Name:    "v0.1.132-1",
			},
		},
		"0.1.132-1",
		"0.1.132",
		"release",
	)

	err := svc.PerformUpdate(context.Background())

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNoUpdateAvailable))
	require.ErrorIs(t, err, ErrNoUpdateAvailable)
}

func TestUpdateServiceCheckUpdateSelectsHighestStableAIWeLinkRelease(t *testing.T) {
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{TagName: "v0.1.170-2"},
		recentReleases: []*GitHubRelease{
			{TagName: "v0.1.170-2.4", Name: "AIWeLink 0.1.170-2.4"},
			{TagName: "v0.1.171", Name: "official-only"},
			{TagName: "v0.1.171-1", Name: "AIWeLink 0.1.171-1"},
			{TagName: "v0.1.172-1", Draft: true},
			{TagName: "v0.1.173-1", Prerelease: true},
			{TagName: "v0.1.999-bad"},
		},
	}
	svc := NewUpdateService(&updateServiceCacheStub{}, client, "0.1.170-2", "0.1.170", "release")

	info, err := svc.CheckUpdate(context.Background(), true)

	require.NoError(t, err)
	require.Equal(t, "AIwelink/sub2api-aiwelink-dev", client.recentRepo)
	require.Empty(t, client.latestRepo)
	require.Equal(t, "0.1.171-1", info.LatestVersion)
	require.Equal(t, "0.1.170", info.UpstreamVersion)
	require.True(t, info.HasUpdate)
}

func newRollbackTestService(current string, releases []*GitHubRelease) *UpdateService {
	return NewUpdateService(
		&updateServiceCacheStub{},
		&updateServiceGitHubClientStub{recentReleases: releases},
		current,
		"0.1.170",
		"release",
	)
}

func TestUpdateServiceListRollbackVersionsFiltersAndCaps(t *testing.T) {
	releases := []*GitHubRelease{
		{TagName: "v0.1.148-1", PublishedAt: "2026-07-09T00:00:00Z"},                   // newer than current: excluded
		{TagName: "v0.1.147-2", PublishedAt: "2026-07-08T00:00:00Z"},                   // current: excluded
		{TagName: "v0.1.147", PublishedAt: "2026-07-07T18:00:00Z"},                     // official-only: excluded
		{TagName: "v0.1.146-2", PublishedAt: "2026-07-07T12:00:00Z", Prerelease: true}, // prerelease: excluded
		{TagName: "v0.1.146-4", PublishedAt: "2026-07-07T00:00:00Z"},
		{TagName: "v0.1.145-1", PublishedAt: "2026-07-06T00:00:00Z", Draft: true}, // draft: excluded
		{TagName: "v0.1.144-3.2", PublishedAt: "2026-07-05T00:00:00Z"},
		{TagName: "v0.1.144-3.2", PublishedAt: "2026-07-05T00:00:00Z"}, // duplicate: excluded
		{TagName: "v0.1.143-9", PublishedAt: "2026-07-04T00:00:00Z"},
		{TagName: "v0.1.142-1", PublishedAt: "2026-07-03T00:00:00Z"}, // beyond cap of 3: excluded
	}
	svc := newRollbackTestService("0.1.147-2", releases)

	versions, err := svc.ListRollbackVersions(context.Background())

	require.NoError(t, err)
	require.Len(t, versions, 3)
	require.Equal(t, "0.1.146-4", versions[0].Version)
	require.Equal(t, "0.1.144-3.2", versions[1].Version)
	require.Equal(t, "0.1.143-9", versions[2].Version)
}

func TestUpdateServiceListRollbackVersionsSortsUnorderedInput(t *testing.T) {
	releases := []*GitHubRelease{
		{TagName: "v0.1.144-1"},
		{TagName: "v0.1.146-1"},
		{TagName: "v0.1.145-1"},
	}
	svc := newRollbackTestService("0.1.147-1", releases)

	versions, err := svc.ListRollbackVersions(context.Background())

	require.NoError(t, err)
	require.Len(t, versions, 3)
	require.Equal(t, "0.1.146-1", versions[0].Version)
	require.Equal(t, "0.1.145-1", versions[1].Version)
	require.Equal(t, "0.1.144-1", versions[2].Version)
}

func TestUpdateServiceListRollbackVersionsEmptyWhenNoneOlder(t *testing.T) {
	releases := []*GitHubRelease{
		{TagName: "v0.1.147-1"},
		{TagName: "v0.1.148-1"},
	}
	svc := newRollbackTestService("0.1.147-1", releases)

	versions, err := svc.ListRollbackVersions(context.Background())

	require.NoError(t, err)
	require.Empty(t, versions)
}

func TestUpdateServiceListRollbackVersionsPropagatesFetchError(t *testing.T) {
	svc := NewUpdateService(
		&updateServiceCacheStub{},
		&updateServiceGitHubClientStub{recentErr: errors.New("github unavailable")},
		"0.1.147-1",
		"0.1.147",
		"release",
	)

	_, err := svc.ListRollbackVersions(context.Background())

	require.Error(t, err)
	require.Contains(t, err.Error(), "github unavailable")
}

func TestUpdateServiceRollbackToVersionRejectsDisallowedTargets(t *testing.T) {
	releases := []*GitHubRelease{
		{TagName: "v0.1.148-1"},
		{TagName: "v0.1.147-2"},
		{TagName: "v0.1.146-1"},
		{TagName: "v0.1.145-1"},
		{TagName: "v0.1.144-1"},
		{TagName: "v0.1.143-1"},
		{TagName: "v0.1.142-1"},
	}
	svc := newRollbackTestService("0.1.147-2", releases)

	for _, target := range []string{
		"",           // empty
		"0.1.147-2",  // current version
		"v0.1.147-2", // current version with prefix
		"0.1.148-1",  // newer than current
		"0.1.142-1",  // older than the 3 most recent
		"9.9.9-1",    // nonexistent
	} {
		err := svc.RollbackToVersion(context.Background(), target)
		require.ErrorIs(t, err, ErrRollbackVersionNotAllowed, "target %q should be rejected", target)
	}
}

func TestUpdateServiceRollbackToVersionAcceptsVPrefix(t *testing.T) {
	// No platform asset in the release: the target passes the allowlist check
	// and fails later at asset lookup, proving the version itself was accepted.
	releases := []*GitHubRelease{
		{TagName: "v0.1.147-2"},
		{TagName: "v0.1.146-1"},
	}
	svc := newRollbackTestService("0.1.147-2", releases)

	err := svc.RollbackToVersion(context.Background(), "v0.1.146-1")

	require.Error(t, err)
	require.NotErrorIs(t, err, ErrRollbackVersionNotAllowed)
	require.Contains(t, err.Error(), "no compatible release found")
}
