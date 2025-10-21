package youtube_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	apiyoutube "google.golang.org/api/youtube/v3"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/youtube"
)

func TestClient_SearchUpcomingLiveStream(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// Given
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/youtube/v3/search", r.URL.Path)
			assert.Equal(t, "snippet", r.URL.Query().Get("part"))
			assert.Equal(t, "chanId", r.URL.Query().Get("channelId"))
			assert.Equal(t, "date", r.URL.Query().Get("order"))
			assert.Equal(t, "video", r.URL.Query().Get("type"))
			assert.Equal(t, "50", r.URL.Query().Get("maxResults"))
			assert.Equal(t, "upcoming", r.URL.Query().Get("eventType"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"items":[{"id":{"videoId":"video1"}},{"id":{"videoId":"video2"}}]}`))
		}

		c := setupTest(t, http.HandlerFunc(handler))

		// When
		upcomingLiveStreamIDs, err := c.SearchUpcomingLiveStream(t.Context(), "chanId")

		// Then
		assert.NoError(t, err)
		assert.ElementsMatch(t, upcomingLiveStreamIDs, []string{"video1", "video2"})
	})

	t.Run("malformed response payload", func(t *testing.T) {
		t.Parallel()

		// Given
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not-json"))
		}

		c := setupTest(t, http.HandlerFunc(handler))

		// When
		upcomingLiveStreamIDs, err := c.SearchUpcomingLiveStream(t.Context(), "chanId")

		// Then
		assert.Nil(t, upcomingLiveStreamIDs)
		assert.ErrorContains(t, err, "call")
	})

	t.Run("non-2xx response", func(t *testing.T) {
		t.Parallel()

		// Given
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		c := setupTest(t, http.HandlerFunc(handler))

		// When
		upcomingLiveStreamIDs, err := c.SearchUpcomingLiveStream(t.Context(), "chanId")

		// Then
		assert.Nil(t, upcomingLiveStreamIDs)
		assert.ErrorContains(t, err, "call")
	})
}

func TestClient_ListLiveStreams(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// Given
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/youtube/v3/videos", r.URL.Path)
			assert.ElementsMatch(t, []string{"video1", "video2"}, r.URL.Query()["id"])
			assert.ElementsMatch(t, []string{"id", "snippet", "liveStreamingDetails"}, r.URL.Query()["part"])
			assert.Equal(t, "50", r.URL.Query().Get("maxResults"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(videoListResponse(t, "video_list/success")))
		}

		c := setupTest(t, http.HandlerFunc(handler))

		// When
		actLiveStreams, err := c.ListLiveStreams(t.Context(), []string{"video1", "video2"})

		// Then
		assert.NoError(t, err)
		assert.Len(t, actLiveStreams, 2)

		now := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)

		ls1, err := domain.NewLiveStream(
			"video1",
			"Test Stream",
			"channel1",
			"Test Channel 1",
			"https://i.ytimg.com/vi/video1/maxresdefault.jpg",
			"live_chat_1",
			now,
			now.Add(12*time.Hour),
		)
		require.NoError(t, err)

		ls2, err := domain.NewLiveStream(
			"video2",
			"Test Stream",
			"channel2",
			"Test Channel 2",
			"https://i.ytimg.com/vi/video2/maxresdefault.jpg",
			"live_chat_2",
			now,
			now.Add(12*time.Hour),
		)
		require.NoError(t, err)

		assert.ElementsMatch(t, []domain.LiveStream{*ls1, *ls2}, actLiveStreams)
	})

	t.Run("malformed response payload", func(t *testing.T) {
		t.Parallel()

		// Given
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not-json"))
		}

		c := setupTest(t, http.HandlerFunc(handler))

		// When
		actLiveStreams, err := c.ListLiveStreams(t.Context(), []string{"whatever"})

		// Then
		assert.Nil(t, actLiveStreams)
		assert.ErrorContains(t, err, "call")
	})

	t.Run("non-200 response", func(t *testing.T) {
		t.Parallel()

		// Given
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		c := setupTest(t, http.HandlerFunc(handler))

		// When
		actLiveStreams, err := c.ListLiveStreams(t.Context(), []string{"whatever"})

		// Then
		assert.Nil(t, actLiveStreams)
		assert.ErrorContains(t, err, "call")
	})

	t.Run("items with missing live stream details or actual end time are skipped", func(t *testing.T) {
		t.Parallel()

		// Given
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(videoListResponse(t, "video_list/skipped")))
		}

		c := setupTest(t, http.HandlerFunc(handler))

		// When
		actLiveStreams, err := c.ListLiveStreams(t.Context(), []string{"whatever"})

		// Then
		assert.NoError(t, err)
		assert.Empty(t, actLiveStreams)
	})

	t.Run("scheduled start time cannot be parsed", func(t *testing.T) {
		t.Parallel()

		// Given
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(videoListResponse(t, "video_list/invalid_start_time")))
		}

		c := setupTest(t, http.HandlerFunc(handler))

		// When
		actLiveStreams, err := c.ListLiveStreams(t.Context(), []string{"whatever"})

		// Then
		assert.ErrorContains(t, err, "parse scheduled start time")
		assert.Nil(t, actLiveStreams)
	})

	t.Run("published time cannot be parsed", func(t *testing.T) {
		t.Parallel()

		// Given
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(videoListResponse(t, "video_list/invalid_published_time")))
		}

		c := setupTest(t, http.HandlerFunc(handler))

		// When
		actLiveStreams, err := c.ListLiveStreams(t.Context(), []string{"whatever"})

		// Then
		assert.ErrorContains(t, err, "parse published at")
		assert.Nil(t, actLiveStreams)
	})

	t.Run("new domain live stream cannot be constructed", func(t *testing.T) {
		t.Parallel()

		// Given
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(videoListResponse(t, "video_list/invalid")))
		}

		c := setupTest(t, http.HandlerFunc(handler))

		// When
		actLiveStreams, err := c.ListLiveStreams(t.Context(), []string{"whatever"})

		// Then
		assert.ErrorContains(t, err, "new live stream")
		assert.Nil(t, actLiveStreams)
	})
}

func setupTest(t *testing.T, handler http.Handler) *youtube.Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	svc, err := apiyoutube.NewService(
		t.Context(),
		option.WithoutAuthentication(),
		option.WithEndpoint(srv.URL),
	)

	require.NotNil(t, svc)
	require.NoError(t, err)

	c, err := youtube.NewClient(svc.Videos, svc.Search)

	require.NotNil(t, c)
	require.NoError(t, err)

	return c
}

func videoListResponse(t *testing.T, path string) []byte {
	t.Helper()

	if path == "" {
		t.FailNow()
	}

	// Get the current test file's directory
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)

	data, err := os.ReadFile(filepath.Clean(filepath.Join(dir, "testdata", path+".json")))
	require.NoError(t, err, "Failed to read test data file")

	return data
}
