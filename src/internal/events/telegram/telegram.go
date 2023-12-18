package telegram

import (
	"bot/LBot/src/internal/clients/telegram"
	"bot/LBot/src/internal/events"
	"bot/LBot/src/internal/storage"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
)

type Processor struct {
	tg *telegram.Client
	offset int
	storage storage.Storage
}

type Meta struct {
	ChatID		int
	Username	string
}

var ErrUnknownEventType = errors.New("unknown event type")

const(
	RndCmd = "/rnd"
	HelpCmd = "/help"
	StartCmd = "/start"
)

func New(client *telegram.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg: client,
		offset: 0,
		storage: storage,
	}
}

func (p *Processor)Fetch(limit int) ([]events.Event, error) {
	const ferr = "events.telegram.Fetch"

	update, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ferr, err)
	}

	if len(update) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(update))

	for _, u := range update {
		res = append(res, event(u))
	}

	p.offset = update[len(update) - 1].ID + 1

	return res, nil
}

func (p *Processor)Process(event events.Event) error {
	const ferr = "events.telegram.Process"

	switch event.Type {
	case events.Message:
		return p.processMessage(event)
	default:
		return fmt.Errorf("%s: %w", ferr, ErrUnknownEventType)
	}
}

func (p *Processor)processMessage(event events.Event) error {
	const ferr = "events.telegram.processMessage"
	
	meta, err := meta(event)
	if err != nil {
		return fmt.Errorf("%s: %w", ferr, err)
	}

	if err := p.doCmd(event.Text, meta.ChatID, meta.Username); err != nil {
		return fmt.Errorf("%s: %w", ferr, err)
	}
	return nil
}

func (p *Processor)doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("command: '%s' | from '%s' | chatID `%d`\n", text, username, chatID)

	if isAddCmd(text) {
		return p.SavePage(chatID, text, username)
	}

	switch text {
	case RndCmd:
		return p.SendRandom(chatID, username)
	case HelpCmd:
		return p.SendHelp(chatID)
	case StartCmd:
		return p.SendHello(chatID)
	default:
		return p.tg.SendMessage(chatID, msgUnknownCommand)
	}
}

func (p *Processor)SavePage(chatID int, pageURL string, username string) error {
	const ferr = "events.telegram.SavePage"

	page := &storage.Page{
		URL: pageURL,
		UserName: username,
	}

	isExists, err := p.storage.IsExists(page)
	if err != nil {
		return fmt.Errorf("%s: %w", ferr, err)
	}
	if isExists {
		return p.tg.SendMessage(chatID, msgAlreadyExists)
	}

	if err := p.storage.Save(page); err != nil {
		return err
	}

	if err := p.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor)SendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor)SendHello(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}

func (p *Processor)SendRandom(chatID int, username string) error {
	const ferr = "events.telegram.SendRandom"

	page, err := p.storage.PickRandom(username)
	if err != nil {
		return fmt.Errorf("%s: %w", ferr, err)
	}

	if err := p.tg.SendMessage(chatID, page.URL); err != nil {
		return fmt.Errorf("%s: %w", ferr, err)
	}

	return p.storage.Remove(page)
}

func isAddCmd(text string) bool {
	u, err := url.Parse(text)
	return err == nil && u.Host != ""
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, errors.New("failed to get meta")
	}
	return res, nil
}

func event(u telegram.Update) events.Event {
	uType := fetchType(u)

	res := events.Event{
		Type: uType,
		Text: fetchText(u),
	}

	if uType == events.Message {
		res.Meta = Meta{
			ChatID: u.Message.Chat.ID,
			Username: u.Message.From.Username,
		}
	}

	return res
}

func fetchType(u telegram.Update) events.Type {
	if u.Message == nil {
		return events.Unknown
	}
	return events.Message
}

func fetchText(u telegram.Update) string {
	if u.Message == nil {
		return ""
	}
	return u.Message.Text
}