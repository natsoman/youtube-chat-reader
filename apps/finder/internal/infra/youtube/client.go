package youtube

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"

	"google.golang.org/api/youtube/v3"
)

type Client struct {
	videoSvc  *youtube.VideosService
	searchSvc *youtube.SearchService
}

func NewClient(videoSvc *youtube.VideosService, searchSvc *youtube.SearchService) (*Client, error) {
	if videoSvc == nil {
		return nil, errors.New("video service is nil")
	}

	if searchSvc == nil {
		return nil, errors.New("search service is nil")
	}

	return &Client{
		videoSvc:  videoSvc,
		searchSvc: searchSvc,
	}, nil
}

func (c *Client) SearchUpcomingLiveStream(ctx context.Context, channelID string) ([]string, error) {
	call := c.searchSvc.List([]string{"snippet"}).
		Context(ctx).
		ChannelId(channelID).
		Order("date").
		Type("video").
		MaxResults(50).
		EventType("upcoming")

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("call: %v", err)
	}

	videoIDs := make([]string, len(resp.Items))

	for i, item := range resp.Items {
		videoIDs[i] = item.Id.VideoId
	}

	return videoIDs, nil
}

func (c *Client) ListLiveStreams(ctx context.Context, videoIDs []string) ([]domain.LiveStream, error) {
	resp, err := c.videoSvc.List([]string{"id", "snippet", "liveStreamingDetails"}).
		Context(ctx).
		Id(videoIDs...).
		MaxResults(50).
		Do()
	if err != nil {
		return nil, fmt.Errorf("call: %v", err)
	}

	liveStreams := make([]domain.LiveStream, 0)

	for _, item := range resp.Items {
		lsd := item.LiveStreamingDetails
		if lsd == nil || lsd.ActualEndTime != "" || lsd.ActiveLiveChatId == "" || lsd.ScheduledStartTime == "" {
			continue
		}

		scheduledStartTime, err := time.Parse(time.RFC3339, item.LiveStreamingDetails.ScheduledStartTime)
		if err != nil {
			return nil, fmt.Errorf("parse scheduled start time: %v", err)
		}

		publishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("parse published at: %v", err)
		}

		ls, err := domain.NewLiveStream(
			item.Id,
			item.Snippet.Title,
			item.Snippet.ChannelId,
			item.Snippet.ChannelTitle,
			item.Snippet.Thumbnails.Maxres.Url,
			item.LiveStreamingDetails.ActiveLiveChatId,
			publishedAt,
			scheduledStartTime,
		)
		if err != nil {
			return nil, fmt.Errorf("new live stream: %v", err)
		}

		liveStreams = append(liveStreams, *ls)
	}

	return liveStreams, nil
}
