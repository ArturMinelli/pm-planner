//go:build windows

package main

import (
	"context"

	toast "git.sr.ht/~jackmordaunt/go-toast/v2"
)

type systemNotifier struct{}

func (systemNotifier) Notify(ctx context.Context, title, body string) error {
	_ = ctx
	_ = toast.SetAppData(toast.AppData{
		AppID: "PM Planner",
		GUID:  "{5DCC0E3A-01F5-45B6-B3EB-499F24CC3D96}",
	})
	notification := toast.Notification{
		AppID: "PM Planner",
		Title: title,
		Body:  body,
	}
	return notification.Push()
}
