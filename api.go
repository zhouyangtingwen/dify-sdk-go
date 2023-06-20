package dify

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	chatMessages = "/v1/chat-messages"
	messagesFeedbacks = "/v1/messages/{message_id}/feedbacks"
	messages = "/v1/messages"
	conversations = "/v1/conversations"
	conversationsRename = "/v1/conversations/{conversation_id}/name"
	parameters = "/v1/parameters"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Params  string `json:"params"`
}

type ChatMessageRequest struct {
	Inputs map[string]interface{} `json:"inputs"`
	Query string `json:"query"`
	ResponseMode string `json:"response_mode"`
	ConversationID string `json:"conversation_id,omitempty"`
	User string `json:"user"`
}

type ChatMessageResponse struct {
	ID string `json:"id"`
	Answer string `json:"answer"`
	ConversationID string `json:"conversation_id"`
	CreatedAt int `json:"created_at"`
}

type ChatMessageStreamResponse struct {
	Event          string `json:"event"`
	TaskID         string `json:"task_id"`
	ID             string `json:"id"`
	Answer         string `json:"answer"`
	CreatedAt      int64  `json:"created_at"`
	ConversationID string `json:"conversation_id"`
}

type ChatMessageStreamChannelResponse struct {
	ChatMessageStreamResponse
	Err error `json:"-"`
}

type MessagesFeedbacksRequest struct {
	MessageID string `json:"message_id,omitempty"`
	Rating string `json:"rating,omitempty"`
	User string `json:"user"`
}

type MessagesFeedbacksResponse struct {
	HasMore bool   `json:"has_more"`
	Data    []MessagesFeedbacksDataResponse `json:"data"`
}

type MessagesFeedbacksDataResponse struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	PhoneNumber    string `json:"phone_number"`
	AvatarURL      string `json:"avatar_url"`
	DisplayName    string `json:"display_name"`
	ConversationID string `json:"conversation_id"`
	LastActiveAt   int64  `json:"last_active_at"`
	CreatedAt      int64  `json:"created_at"`
}

type MessagesRequest struct {
	ConversationID string `json:"conversation_id"`
	FirstID string `json:"first_id,omitempty"`
	Limit int `json:"limit"`
	User string `json:"user"`
}

type MessagesResponse struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	PhoneNumber    string `json:"phone_number"`
	AvatarURL      string `json:"avatar_url"`
	DisplayName    string `json:"display_name"`
	ConversationID string `json:"conversation_id"`
	LastActiveAt   int64  `json:"last_active_at"`
	CreatedAt      int64  `json:"created_at"`
}

type ConversationsRequest struct {
	LastID string `json:"last_id,omitempty"`
	Limit int `json:"limit"`
	User string `json:"user"`
}

type ConversationsResponse struct {
	Limit   int    `json:"limit"`
	HasMore bool   `json:"has_more"`
	Data    []ConversationsDataResponse `json:"data"`
}

type ConversationsDataResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Inputs    map[string]string `json:"inputs"`
	Status    string            `json:"status"`
	CreatedAt int64             `json:"created_at"`
}

type ConversationsRenamingRequest struct {
	ConversationID string `json:"conversation_id,omitempty"`
	Name string `json:"name"`
	User string `json:"user"`
}

type ConversationsRenamingResponse struct {
	Result string `json:"result"`
}

type ParametersRequest struct {
	User string `json:"user"`
}

type ParametersResponse struct {
	Text      string      `json:"introduction"`
	Variables []ParametersVariableResponse  `json:"variables"`
}

type ParametersVariableResponse struct {
	Key         string      `json:"key"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Default     interface{} `json:"default"`
	Options     interface{} `json:"options"`
}

const (
	FeedbackLike = "like"
	FeedbackDislike = "dislike"
)

type Api struct {
	c *Client
}

func (api *Api) buildRequestApi(requestUrl string) string {
	return api.c.GetHost() + requestUrl
}

func (api *Api) createBaseRequest(ctx context.Context, method string, url string, req ...interface{}) (r *http.Request, err error) {
	if r, err = api.c.NewHttpRequest(ctx, http.MethodPost, url, req...); err == nil {
		api.c.SetHttpRequest(r).
			SetHttpRequestHeader("Authorization", "Bearer " + api.c.GetApiSecretKey()).
			SetHttpRequestHeader("Cache-Control", "no-cache").
			SetHttpRequestHeader("Content-Type", "application/json; charset=utf-8")
	}
	return
}

func (api *Api) createGetRequest(ctx context.Context, url string, req ...interface{}) (*http.Request, error) {
	return api.createBaseRequest(ctx, http.MethodGet, url, req...)
}

func (api *Api) createPostRequest(ctx context.Context, url string, req ...interface{}) (*http.Request, error) {
	return api.createBaseRequest(ctx, http.MethodPost, url, req...)
}

func (api *Api) createChatMessageHttpRequest(ctx context.Context, req *ChatMessageRequest) (r *http.Request, err error) {
	r, err = api.createPostRequest(ctx, api.buildRequestApi(chatMessages), req)
	return
}

/* Create chat message
 * Create a new conversation message or continue an existing dialogue.
 */
func (api *Api) ChatMessages(ctx context.Context, req *ChatMessageRequest) (resp *ChatMessageResponse, err error) {
	if req == nil {
		err = errors.New("ChatMessages.ChatMessageRequest Illegal")
		return
	}
	req.ResponseMode = "blocking"

	var r *http.Request
	if r, err = api.createChatMessageHttpRequest(ctx, req); err != nil {
		return
	}

	var _resp ChatMessageResponse

	err = api.c.SetHttpRequest(r).SendRequest(&_resp)
	resp = &_resp
	return
}

func (api *Api) ChatMessagesStreamRaw(ctx context.Context, req *ChatMessageRequest) (resp *http.Response, err error) {
	if req == nil {
		err = errors.New("ChatMessagesStreamRaw.ChatMessageRequest Illegal")
		return
	}
	req.ResponseMode = "streaming"

	var r *http.Request
	if r, err = api.createChatMessageHttpRequest(ctx, req); err != nil {
		return
	}

	return api.c.SetHttpRequest(r).SendRequestStream()
}

func (api *Api) ChatMessagesStream(ctx context.Context, req *ChatMessageRequest) (streamChannel chan ChatMessageStreamChannelResponse, err error) {
	if req == nil {
		err = errors.New("ChatMessagesStream.ChatMessageRequest Illegal")
		return
	}

	var resp *http.Response
	if resp, err = api.ChatMessagesStreamRaw(ctx, req); err != nil {
		return
	}

	streamChannel = make(chan ChatMessageStreamChannelResponse)
	go api.chatMessagesStreamHandle(ctx, resp, streamChannel)
	return
}

func (api *Api) chatMessagesStreamHandle(ctx context.Context, resp *http.Response, streamChannel chan ChatMessageStreamChannelResponse) {
	var (
		body = resp.Body
		reader = bufio.NewReader(body)

		err error
		line []byte
	)

	defer resp.Body.Close()
	defer close(streamChannel)

	if line, _, err = reader.ReadLine(); err == nil {
		var errResp ErrorResponse
		var _err error
		if _err = json.Unmarshal(line, &errResp); _err == nil {
			streamChannel <-ChatMessageStreamChannelResponse{
				Err: errors.New(string(line)),
			}
			return
		}
	}

	for {
		select {
			case <-ctx.Done():
				return
			default:
				if line, err = reader.ReadBytes('\n'); err != nil {
					streamChannel <-ChatMessageStreamChannelResponse{
						Err: errors.New("Error reading line: " + err.Error()),
					}
					return
				}

				if !bytes.HasPrefix(line, []byte("data:")) {
					continue
				}

				line = bytes.TrimPrefix(line, []byte("data:"))
				line = bytes.TrimSpace(line)

				var resp ChatMessageStreamChannelResponse
				if err = json.Unmarshal(line, &resp); err != nil {
					streamChannel <-ChatMessageStreamChannelResponse{
						Err: errors.New("Error unmarshalling event: " + err.Error()),
					}
					return
				} else if resp.Answer == "" {
					return
				}
				streamChannel <-resp
		}
	}
}

/* Message terminal user feedback, like
 * Rate received messages on behalf of end-users with likes or dislikes. 
 * This data is visible in the Logs & Annotations page and used for future model fine-tuning.
 */
func (api *Api) MessagesFeedbacks(ctx context.Context, req *MessagesFeedbacksRequest) (resp *MessagesFeedbacksResponse, err error) {
	if req == nil {
		err = errors.New("MessagesFeedbacks.MessagesFeedbacksRequest Illegal")
		return
	}
	if req.MessageID == "" {
		err = errors.New("MessagesFeedbacksRequest.MessageID Illegal")
		return
	}

	var url = api.buildRequestApi(messagesFeedbacks)
	url = strings.ReplaceAll(url, "{message_id}", req.MessageID)

	req.MessageID = ""

	var r *http.Request
	if r, err = api.createPostRequest(ctx, url, req); err != nil {
		return
	}

	var _resp MessagesFeedbacksResponse

	err = api.c.SetHttpRequest(r).SendRequest(&_resp)
	resp = &_resp
	return
}

/* Get the chat history message
 * The first page returns the latest limit bar, which is in reverse order.
 */
func (api *Api) Messages(ctx context.Context, req *MessagesRequest) (resp *MessagesResponse, err error) {
	if req == nil {
		err = errors.New("Messages.MessagesRequest Illegal")
		return
	}

	var u = url.Values{}
	u.Set("conversation_id", req.ConversationID)
	u.Set("user", req.User)

	if req.FirstID != "" {
		u.Set("first_id", req.FirstID)
	}
	if req.Limit > 0 {
		var l = int64(req.Limit)
		u.Set("limit", strconv.FormatInt(l, 10))
	}

	var r *http.Request
	if r, err = api.createGetRequest(ctx, api.buildRequestApi(messages), u); err != nil {
		return
	}

	var _resp MessagesResponse

	err = api.c.SetHttpRequest(r).SendRequest(&_resp)
	resp = &_resp
	return
}

/* Get conversation list
 * Gets the session list of the current user. By default, the last 20 sessions are returned.
 */
func (api *Api) Conversations(ctx context.Context, req *ConversationsRequest) (resp *ConversationsResponse, err error) {
	if req == nil {
		err = errors.New("Conversations.ConversationsRequest Illegal")
		return
	}

	if req.User == "" {
		err = errors.New("ConversationsRequest.User Illegal")
		return
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	var u = url.Values{}
	u.Set("last_id", req.LastID)
	u.Set("user", req.User)

	var l = int64(req.Limit)
	u.Set("limit", strconv.FormatInt(l, 10))

	var r *http.Request
	if r, err = api.createGetRequest(ctx, api.buildRequestApi(conversations), u); err != nil {
		return
	}

	var _resp ConversationsResponse

	err = api.c.SetHttpRequest(r).SendRequest(&_resp)
	resp = &_resp
	return
}

/* Conversation renaming
 * Rename conversations; the name is displayed in multi-session client interfaces.
 */
func (api *Api) ConversationsRenaming(ctx context.Context, req *ConversationsRenamingRequest) (resp *ConversationsRenamingResponse, err error) {
	if req == nil {
		err = errors.New("ConversationsRenaming.ConversationsRenamingRequest Illegal")
		return
	}

	var url = api.buildRequestApi(conversationsRename)
	url = strings.ReplaceAll(url, "{conversation_id}", req.ConversationID)

	req.ConversationID = ""

	var r *http.Request
	if r, err = api.createPostRequest(ctx, url, req); err != nil {
		return
	}

	var _resp ConversationsRenamingResponse

	err = api.c.SetHttpRequest(r).SendRequest(&_resp)
	resp = &_resp
	return
}

/* Obtain application parameter information
 * Retrieve configured Input parameters, including variable names, field names, types, and default values. 
 * Typically used for displaying these fields in a form or filling in default values after the client loads.
 */
func (api *Api) Parameters(ctx context.Context, req *ParametersRequest) (resp *ParametersResponse, err error) {
	if req == nil {
		err = errors.New("Parameters.ParametersRequest Illegal")
		return
	}
	if req.User == "" {
		err = errors.New("ParametersRequest.User Illegal")
		return
	}

	var u = url.Values{}
	u.Set("user", req.User)

	var r *http.Request
	if r, err = api.createGetRequest(ctx, api.buildRequestApi(parameters), u); err != nil {
		return
	}

	var _resp ParametersResponse

	err = api.c.SetHttpRequest(r).SendRequest(&_resp)
	resp = &_resp
	return
}