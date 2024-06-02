// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
)

const (
	receiveChanSize = 256
)

// Subcription represents a topic subscription to the websocket Hub interface.
type Subcription struct {
	sid int

	topic  Topic
	params Parameters

	mch chan *WrapperObject
}

type SubcriptionOpt func(*Subcription) error

func WithSubcriptionParams(params ...Parameter) SubcriptionOpt {
	return func(s *Subcription) error {
		s.params = make(Parameters)
		for _, param := range params {
			if err := param(s.topic, s.params); err != nil {
				return err
			}
		}
		return nil
	}
}

func NewSubscription(topic Topic, opts ...SubcriptionOpt) (*Subcription, error) {
	sub := Subcription{
		topic: topic,
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(&sub); err != nil {
				return nil, err
			}
		}
	}

	return &sub, nil
}

func ReceiveAs[T any](sub *Subcription) <-chan *T {
	out := make(chan *T, receiveChanSize)

	go func() {
		defer close(out)

		for msg := range sub.mch {
			for _, payload := range msg.Payload {
				var v T
				if err := json.Unmarshal(payload, &v); err != nil {
					continue
				}

				out <- &v
			}
		}
	}()

	return out
}

func (s Subcription) ReceiveRaw() <-chan *WrapperObject {
	return s.mch
}

func (s *Subcription) close() {
	close(s.mch)
}
