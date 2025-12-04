package airbus

import (
	"context"
	"net/http"
)

// WhoAmI returns account information for the current user.
// GET /user/whoami
func (c *Client) WhoAmI(ctx context.Context) (*UserInfo, error) {
	var out UserInfo
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("user", "whoami"), nil, http.StatusOK, &out)
	return &out, err
}

// ChangePassword changes the account password.
// The oldPassword and password fields should be base64 encoded.
// POST /user/password
func (c *Client) ChangePassword(ctx context.Context, req *ChangePasswordRequest) error {
	body, err := marshalBody(req)
	if err != nil {
		return err
	}
	return c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("user", "password"), body, http.StatusOK, nil)
}

// ResetPassword requests a password reset for the given username.
// POST /user/password/reset
func (c *Client) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	body, err := marshalBody(req)
	if err != nil {
		return err
	}
	return c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("user", "password", "reset"), body, http.StatusOK, nil)
}

// ListNotifications retrieves all notifications for the current user.
// GET /user/notifications
func (c *Client) ListNotifications(ctx context.Context) ([]Notification, error) {
	var out []Notification
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("user", "notifications"), nil, http.StatusOK, &out)
	return out, err
}

// UpdateNotification updates a notification (e.g., mark as read).
// PATCH /user/notifications/{notificationId}
func (c *Client) UpdateNotification(ctx context.Context, notificationID string, req *UpdateNotificationRequest) error {
	body, err := marshalBody(req)
	if err != nil {
		return err
	}
	return c.doRequest(ctx, http.MethodPatch, c.BaseURL().JoinPath("user", "notifications", notificationID), body, http.StatusOK, nil)
}

// DeleteNotification deletes a notification.
// Maintenance notifications cannot be deleted.
// DELETE /user/notifications/{notificationId}
func (c *Client) DeleteNotification(ctx context.Context, notificationID string) error {
	return c.doRequest(ctx, http.MethodDelete, c.BaseURL().JoinPath("user", "notifications", notificationID), nil, http.StatusNoContent, nil)
}
