// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package client

import (
	"fmt"
	"strconv"

	"golang.org/x/net/context"

	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/protocol/chat1"
	"github.com/keybase/client/go/protocol/keybase1"
)

type chatConversationResolver struct {
	TlfName    string
	TopicName  string
	TopicType  chat1.TopicType
	Visibility chat1.TLFVisibility
}

func (r *chatConversationResolver) Resolve(ctx context.Context, g *libkb.GlobalContext, chatClient chat1.LocalInterface, tlfClient keybase1.TlfInterface) (conversationInfo *chat1.ConversationInfoLocal, userChosen bool, err error) {
	if len(r.TlfName) > 0 {
		cname, err := tlfClient.CompleteAndCanonicalizeTlfName(ctx, r.TlfName)
		if err != nil {
			return nil, false, fmt.Errorf("completing TLF name error: %v", err)
		}
		r.TlfName = string(cname)
	}

	rcres, err := chatClient.ResolveConversationLocal(ctx, chat1.ConversationInfoLocal{
		TlfName:    r.TlfName,
		TopicName:  r.TopicName,
		TopicType:  r.TopicType,
		Visibility: r.Visibility,
	})
	if err != nil {
		return nil, false, err
	}

	conversations := rcres.Convs
	switch len(conversations) {
	case 0:
		return nil, false, nil
	case 1:
		return &conversations[0], false, nil
	default:
		g.UI.GetTerminalUI().Printf(
			"There are %d conversations. Please choose one:\n", len(conversations))
		conversationInfoListView(conversations).show(g)
		var num int
		for num = -1; num < 1 || num > len(conversations); {
			input, err := g.UI.GetTerminalUI().Prompt(PromptDescriptorChooseConversation,
				fmt.Sprintf("Please enter a number [1-%d]: ", len(conversations)))
			if err != nil {
				return nil, false, err
			}
			if num, err = strconv.Atoi(input); err != nil {
				g.UI.GetTerminalUI().Printf("Error converting input to number: %v\n", err)
				continue
			}
		}
		return &conversations[num-1], true, nil
	}
}

type chatConversationFetcher struct {
	selector chat1.MessageSelector
	resolver chatConversationResolver

	chatClient chat1.LocalInterface // for testing only
}

func (f chatConversationFetcher) fetch(ctx context.Context, g *libkb.GlobalContext) (conversations []chat1.ConversationLocal, err error) {
	chatClient := f.chatClient // should be nil unless in test
	if chatClient == nil {
		chatClient, err = GetChatLocalClient(g)
		if err != nil {
			return nil, fmt.Errorf("Getting chat service client error: %s", err)
		}
	}

	tlfClient, err := GetTlfClient(g)
	if err != nil {
		return nil, err
	}

	conversationInfo, _, err := f.resolver.Resolve(ctx, g, chatClient, tlfClient)
	if err != nil {
		return nil, fmt.Errorf("resolving conversation error: %v\n", err)
	}
	if conversationInfo == nil {
		return nil, nil
	}
	g.UI.GetTerminalUI().Printf("fetching conversation %s ...\n", conversationInfo.TlfName)
	f.selector.Conversations = append(f.selector.Conversations, conversationInfo.Id)

	gmres, err := chatClient.GetMessagesLocal(ctx, f.selector)
	if err != nil {
		return nil, fmt.Errorf("GetMessagesLocal error: %s", err)
	}

	return gmres.Msgs, nil
}

type chatInboxFetcher struct {
	query chat1.GetInboxSummaryLocalQuery

	chatClient chat1.LocalInterface // for testing only
}

func (f chatInboxFetcher) fetch(ctx context.Context, g *libkb.GlobalContext) (conversations []chat1.ConversationLocal, more []chat1.ConversationLocal, moreTotal int, err error) {
	chatClient := f.chatClient // should be nil unless in test
	if chatClient == nil {
		chatClient, err = GetChatLocalClient(g)
		if err != nil {
			return nil, nil, moreTotal, fmt.Errorf("Getting chat service client error: %s", err)
		}
	}

	res, err := chatClient.GetInboxSummaryLocal(ctx, f.query)
	if err != nil {
		return nil, nil, moreTotal, err
	}

	return res.Conversations, res.More, res.MoreTotal, nil
}
