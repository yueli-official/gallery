package galleryabuse

import (
	"net/netip"
	"strings"
	"time"

	"github.com/yueli-official/foundation/go/abuse"
)

const (
	ActionGuestSubmission  abuse.ActionKey = "gallery.submission.create.guest"
	ActionMemberSubmission abuse.ActionKey = "gallery.submission.create.member"
	ActionGuestComment     abuse.ActionKey = "gallery.comment.create.guest"
	ActionMemberComment    abuse.ActionKey = "gallery.comment.create.member"
)

type Policy struct {
	Challenge *abuse.ChallengeDefinition
}

func Definition(policy Policy) abuse.Definition {
	guestChallengeAt := int64(0)
	if policy.Challenge != nil {
		guestChallengeAt = 5
	}
	return abuse.Definition{
		Version:  2,
		Consumer: "gallery",
		Actions: []abuse.ActionDefinition{
			{
				Key: ActionGuestSubmission,
				Required: abuse.SignalRequirements{
					Network: abuse.Required,
					Actor:   abuse.Required,
				},
				Meters: []abuse.MeterDefinition{
					{
						ID:        "gallery.submission.guest.network",
						Slot:      abuse.SlotNetwork,
						Algorithm: abuse.TokenBucket(10, 10, time.Hour),
					},
					{
						ID:          "gallery.submission.guest.actor",
						Slot:        abuse.SlotActor,
						Algorithm:   abuse.SlidingWindow(5, time.Hour),
						ChallengeAt: guestChallengeAt,
					},
				},
				Challenge: policy.Challenge,
			},
			{
				Key: ActionMemberSubmission,
				Required: abuse.SignalRequirements{
					Network: abuse.Required,
					Actor:   abuse.Required,
				},
				Meters: []abuse.MeterDefinition{
					{
						ID:        "gallery.submission.member.network",
						Slot:      abuse.SlotNetwork,
						Algorithm: abuse.TokenBucket(60, 60, time.Hour),
					},
					{
						ID:        "gallery.submission.member.actor",
						Slot:      abuse.SlotActor,
						Algorithm: abuse.SlidingWindow(20, 24*time.Hour),
					},
				},
			},
			{
				Key: ActionGuestComment,
				Required: abuse.SignalRequirements{
					Network: abuse.Required,
					Actor:   abuse.Required,
				},
				Meters: []abuse.MeterDefinition{
					{ID: "gallery.comment.guest.network", Slot: abuse.SlotNetwork, Algorithm: abuse.TokenBucket(20, 20, time.Hour)},
					{ID: "gallery.comment.guest.actor", Slot: abuse.SlotActor, Algorithm: abuse.SlidingWindow(8, time.Hour)},
				},
			},
			{
				Key: ActionMemberComment,
				Required: abuse.SignalRequirements{
					Network: abuse.Required,
					Actor:   abuse.Required,
				},
				Meters: []abuse.MeterDefinition{
					{ID: "gallery.comment.member.network", Slot: abuse.SlotNetwork, Algorithm: abuse.TokenBucket(120, 120, time.Hour)},
					{ID: "gallery.comment.member.actor", Slot: abuse.SlotActor, Algorithm: abuse.SlidingWindow(40, 24*time.Hour)},
				},
			},
		},
	}
}

type Actions struct {
	Guest         abuse.Action
	Member        abuse.Action
	GuestComment  abuse.Action
	MemberComment abuse.Action
}

func Bind(module abuse.Module) (Actions, error) {
	guest, err := module.Action(ActionGuestSubmission)
	if err != nil {
		return Actions{}, err
	}
	member, err := module.Action(ActionMemberSubmission)
	if err != nil {
		return Actions{}, err
	}
	guestComment, err := module.Action(ActionGuestComment)
	if err != nil {
		return Actions{}, err
	}
	memberComment, err := module.Action(ActionMemberComment)
	if err != nil {
		return Actions{}, err
	}
	return Actions{Guest: guest, Member: member, GuestComment: guestComment, MemberComment: memberComment}, nil
}

func NetworkPrefix(value string) (netip.Prefix, error) {
	address, err := netip.ParseAddr(strings.TrimSpace(value))
	if err != nil {
		return netip.Prefix{}, err
	}
	address = address.Unmap()
	bits := address.BitLen()
	if address.Is6() {
		bits = 64
	}
	return netip.PrefixFrom(address, bits).Masked(), nil
}
