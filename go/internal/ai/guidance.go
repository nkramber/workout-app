package ai

// GuidanceKind names the kind of a guidance item (D-44).
type GuidanceKind string

const (
	KindWarmUp   GuidanceKind = "warm_up"
	KindCoolDown GuidanceKind = "cool_down"
	KindMobility GuidanceKind = "mobility"
	KindRecovery GuidanceKind = "recovery"
)

// GuidanceID is the stable id of a guidance item. An id never changes,
// and a removed id is never used again.
type GuidanceID string

// GuidanceItem is one short text of the guidance catalog (D-152). Luna
// selects an item by id and writes no prose for it. The texts stay
// inside the fitness boundary of D-36, and a code review checks each
// change.
type GuidanceItem struct {
	ID   GuidanceID
	Kind GuidanceKind
	Text string
}

// GuidanceVersion is the version of the guidance catalog. Change it
// with each change of the items.
const GuidanceVersion = 1

// The items that a session gets when it has no valid selection.
const (
	DefaultWarmUp   GuidanceID = "warm_up.easy_cardio"
	DefaultCoolDown GuidanceID = "cool_down.easy_walk"
)

var guidance = []GuidanceItem{
	{"warm_up.easy_cardio", KindWarmUp, "Do 5 minutes of easy cardio. Then do one light set of the first exercise."},
	{"warm_up.light_sets", KindWarmUp, "Do one light set before the first working set of each exercise."},
	{"cool_down.easy_walk", KindCoolDown, "Walk at an easy pace for 5 minutes."},
	{"cool_down.stretch", KindCoolDown, "Stretch the muscles that you trained. Hold each stretch for about 30 seconds, and stop before it hurts."},
	{"mobility.hips", KindMobility, "Do 10 slow leg swings to the front and 10 to the side on each leg."},
	{"mobility.shoulders", KindMobility, "Do 10 slow arm circles to the front and 10 to the back."},
	{"mobility.upper_back", KindMobility, "Sit tall, and turn your upper body slowly to each side 10 times."},
	{"recovery.rest_day", KindRecovery, "Leave one rest day between two sessions that train the same muscles."},
	{"recovery.easy_walk", KindRecovery, "On a rest day, an easy walk keeps you moving."},
	{"recovery.sleep", KindRecovery, "Keep a regular sleep time on training days."},
}

// Guidance gives a new copy of each item of the catalog, in a fixed
// order.
func Guidance() []GuidanceItem { return append([]GuidanceItem(nil), guidance...) }

// GuidanceItemOf gives the item of an id.
func GuidanceItemOf(id GuidanceID) (GuidanceItem, bool) {
	for _, g := range guidance {
		if g.ID == id {
			return g, true
		}
	}
	return GuidanceItem{}, false
}

func guidanceIDs(kinds ...GuidanceKind) []GuidanceID {
	var out []GuidanceID
	for _, g := range guidance {
		for _, k := range kinds {
			if g.Kind == k {
				out = append(out, g.ID)
			}
		}
	}
	return out
}
