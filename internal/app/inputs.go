package app

import (
	"context"

	"ghgo/internal/input"
	"ghgo/internal/store"
	"ghgo/internal/vocab"
)

type InputService struct {
	store *store.Store
}

type ParseInputOptions struct {
	InputKind vocab.InputKind
	RawText   string
}

type CommitParsedInputOptions struct {
	Context input.CommitContext
	Parsed  input.ParseResult
}

type ParseAndCommitInputOptions struct {
	Context input.CommitContext
	RawText string
}

func (s *InputService) Parse(ctx context.Context, opts ParseInputOptions) (input.ParseResult, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return input.ParseResult{}, err
	}
	if !opts.InputKind.Valid() {
		return input.ParseResult{}, invalidOptions("invalid input kind %q", opts.InputKind)
	}
	return input.Parse(opts.InputKind, opts.RawText), nil
}

func (s *InputService) CommitParsed(ctx context.Context, opts CommitParsedInputOptions) (input.CommitResult, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return input.CommitResult{}, err
	}
	return input.CommitParsedInput(ctx, s.store, opts.Context, opts.Parsed)
}

func (s *InputService) ParseAndCommit(ctx context.Context, opts ParseAndCommitInputOptions) (input.CommitResult, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return input.CommitResult{}, err
	}
	parsed := input.Parse(opts.Context.InputKind, opts.RawText)
	return s.CommitParsed(ctx, CommitParsedInputOptions{Context: opts.Context, Parsed: parsed})
}
