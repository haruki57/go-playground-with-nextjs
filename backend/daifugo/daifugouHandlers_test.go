package daifugo

import (
	"maps"
	"testing"
)


func Test_game(t *testing.T) {
	game := createGameWithStandardRules()
	err := game.startGame()
	if err == nil {
		t.Errorf("should be error")
	}
	for _, playerName := range []string{"p1", "p2", "p3", "p4", "p5",} {
		err = game.addPlayer(playerName)
		if err != nil {
			t.Errorf("should not be error")
		}
	}
	game.removePlayer("p5")
	err = game.startGame()
	if err != nil {
		t.Errorf("should not be error")
	}
	if len(game.Players) != 4 {
		t.Errorf("num of players should be 4")
	}
	for _, player := range game.Players {
		if player.Role != Heimin {
			t.Errorf("all roles should be Heimin")
		}
		if !(13 <= len(player.Cards) && len(player.Cards) <= 14) {
			t.Errorf("num of cards should be either 13 or 14")
		}
	}
	if game.GameState != PlayingCards {
		t.Errorf("GameState should be PlayingCards")
	}
	if game.Turn != 0 {
		t.Errorf("turn should be 0")
	}
	
	submitted := game.submitCards(game.Players[1], []Card{game.Players[1].Cards[0]})
	if submitted {
		t.Errorf("submitted should be false")
	}
	submitted = game.submitCards(game.Players[0], []Card{game.Players[0].Cards[0]})
	if !submitted {
		t.Errorf("submitted should be true")
	}
	if len(game.PlayingCards) != 1 {
		t.Errorf("len(game.PlayingCards) should be 1")
	}
	if game.Turn != 1 {
		t.Errorf("Turn should be 1")
	}
	game.pass(game.Players[1])
	if game.Turn != 2 {
		t.Errorf("Turn should be 2")
	}
	game.pass(game.Players[2])
	game.pass(game.Players[3])
	if game.Turn != 0 {
		t.Errorf("Turn should be 0")
	}
	if len(game.PlayingCards) != 0 {
		t.Errorf("len(game.PlayingCards) should be 0")
	}
	// submitcards を、game.submitCards(player, cards)にするのか、player.submitCards(game, cards)にするのか？ん〜？
	// 
}

func Test_canSubmitCards(t *testing.T) {
	type args struct {
		topFieldCards, submittingCards []Card
		submitModes map[SubmitMode]struct{}
		specialRules map[SpecialRule]struct{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty vs 4", args{[]Card{}, []Card{makeCard(4, Diamond)}, 
			nil, maps.Clone(StandardRule)}, true},
		{"3 vs 4", args{[]Card{makeCard(3, Spade)}, []Card{makeCard(4, Diamond)}, 
			nil, maps.Clone(StandardRule)}, true},
		{"4 vs 4", args{[]Card{makeCard(4, Spade)}, []Card{makeCard(4, Diamond)}, 
			nil, maps.Clone(StandardRule)}, false},
		{"2 vs 3", args{[]Card{makeCard(2, Spade)}, []Card{makeCard(3, Diamond)}, 
			nil, maps.Clone(StandardRule)}, false},
		{"3 vs 4 under kakumai", args{[]Card{makeCard(3, Spade)}, []Card{makeCard(4, Diamond)}, 
			map[SubmitMode]struct{}{KakumeiMode: {}}, maps.Clone(StandardRule)}, false},
		{"Joker vs Spade 3", args{[]Card{makeCard(99, Joker)}, []Card{makeCard(3, Spade)}, 
			nil, maps.Clone(StandardRule)}, true},
		{"3_3 vs 4_4", args{
			[]Card{makeCard(3, Spade), makeCard(3, Diamond)}, 
			[]Card{makeCard(4, Diamond), makeCard(4, Heart)}, 
			nil, maps.Clone(StandardRule)}, true},
		{"3_3 vs 4_5", args{
			[]Card{makeCard(3, Spade), makeCard(3, Diamond)}, 
			[]Card{makeCard(4, Diamond), makeCard(5, Heart)}, 
			nil, maps.Clone(StandardRule)}, false},
		{"3_3 vs 4_5", args{
			[]Card{makeCard(3, Spade), makeCard(3, Diamond)}, 
			[]Card{makeCard(4, Diamond), makeCard(5, Heart)}, 
			nil, maps.Clone(StandardRule)}, false},				
		{"3_3 vs 4_Joker", args{
			[]Card{makeCard(3, Spade), makeCard(3, Diamond)}, 
			[]Card{makeCard(4, Diamond), makeCard(-1, Joker)}, 
			nil, maps.Clone(StandardRule)}, true},				
		{"2 vs joker", args{[]Card{makeCard(2, Spade)}, []Card{makeCard(-1, Joker)}, 
			nil, maps.Clone(StandardRule)}, true},
		{"2 vs joker under kakumai", args{[]Card{makeCard(2, Spade)}, []Card{makeCard(-1, Joker)}, 
			map[SubmitMode]struct{}{KakumeiMode: {}}, maps.Clone(StandardRule)}, true},
		{"2_2 vs joker_joker", args{
			[]Card{makeCard(2, Spade), makeCard(2, Diamond)},
		  []Card{makeCard(-1, Joker), makeCard(-1, Joker)},
			nil, maps.Clone(StandardRule)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, reason := canSubmitCards(tt.args.topFieldCards, tt.args.submittingCards, tt.args.submitModes, tt.args.specialRules); got != tt.want {
				t.Errorf("%s failed with reason:'%s'. got %v, want %v", tt.name, reason, got, tt.want)
			}
		})
	}
}