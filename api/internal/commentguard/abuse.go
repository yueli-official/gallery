// Package commentguard adapts Gallery's Foundation Abuse actions to the Comment Module guard seam.
package commentguard

import (
	"context"
	"strings"

	"github.com/yueli-official/foundation/go/abuse"
	"github.com/yueli-official/gallery/api/internal/galleryabuse"
	"github.com/yueli-official/gallery/api/internal/gallerycomments"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
)

type Abuse struct {
	guest  abuse.Action
	member abuse.Action
}

func New(actions galleryabuse.Actions) *Abuse {
	return &Abuse{guest: actions.GuestComment, member: actions.MemberComment}
}

func (guard *Abuse) Admit(ctx context.Context, actor gallerycomments.Actor, attemptID, proof string) error {
	action := guard.guest
	actorKey := "guest:" + strings.TrimSpace(actor.IP)
	if strings.TrimSpace(actor.UserKey) != "" {
		action = guard.member
		actorKey = "user:" + strings.TrimSpace(actor.UserKey)
	}
	if action == nil {
		return nil
	}
	network, err := galleryabuse.NetworkPrefix(actor.IP)
	if err != nil {
		return galleryerr.AbuseUnavailable()
	}
	input := abuse.Input{
		ID:      abuse.AttemptID(strings.TrimSpace(attemptID)),
		Signals: abuse.Signals{Network: network, Actor: actorKey},
	}
	if strings.TrimSpace(proof) != "" {
		input.Proof = &abuse.Proof{Kind: "turnstile", Token: strings.TrimSpace(proof)}
	}
	admission, err := action.Admit(ctx, input)
	if err != nil {
		if abuse.IsKind(err, abuse.ErrorConflict) {
			return galleryerr.AbuseAttemptReplayed()
		}
		return galleryerr.AbuseUnavailable()
	}
	switch admission.Disposition {
	case abuse.DispositionAllow:
		if admission.Replay {
			return galleryerr.AbuseAttemptReplayed()
		}
		return nil
	case abuse.DispositionChallenge:
		return galleryerr.ChallengeRequired(attemptID)
	default:
		return galleryerr.RateLimited()
	}
}
