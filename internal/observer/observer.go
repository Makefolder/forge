// Forge - Automated Docker container deployment tool for VPS environments.
// Monitors Git repositories and redeploys containers on new commits.
// Copyright (C) 2025 Artemii Fedotov
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
package observer

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"smithery/forge/internal/clients/git"
	"time"
)

var lastPushed time.Time = time.Now()

type IObserver interface {
	Observe(ctx context.Context, u *url.URL) error
}

type Observer struct {
	subscriptions []func(context.Context) error
	git           git.IGitClient
	interval      time.Duration
}

type ObserverParams struct {
	Git           git.IGitClient
	Interval      time.Duration
	Subscriptions []func(context.Context) error
}

func New(params ObserverParams) IObserver {
	return &Observer{
		git:           params.Git,
		interval:      params.Interval,
		subscriptions: params.Subscriptions,
	}
}

func (o *Observer) Observe(ctx context.Context, u *url.URL) error {
	if u == nil {
		return errors.New("URL cannot be nil")
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			r, err := o.git.GetRepository(ctx)
			if err != nil {
				return err
			}
			if r.PushedAt.After(lastPushed) {
				o.notify(ctx)
			}
			time.Sleep(o.interval)
		}
	}
}

func (o *Observer) notify(ctx context.Context) {
	if len(o.subscriptions) == 0 {
		return
	}
	for idx, sub := range o.subscriptions {
		if err := sub(ctx); err != nil {
			slog.Error("Failed to notify", slog.Int("idx", idx), "error", err)
		}
	}
}
