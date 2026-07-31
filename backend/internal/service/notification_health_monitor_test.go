package service

import "testing"

func TestPlanNotificationHealthAlerts(t *testing.T) {
	tests := []struct {
		name       string
		tgMon      bool
		tgDown     bool
		maxMon     bool
		maxDown    bool
		wantEmail  bool
		wantTG     bool
		wantMax    bool
	}{
		{
			name: "both healthy",
			tgMon: true, tgDown: false, maxMon: true, maxDown: false,
		},
		{
			name: "telegram down max healthy",
			tgMon: true, tgDown: true, maxMon: true, maxDown: false,
			wantEmail: true, wantMax: true,
		},
		{
			name: "max down telegram healthy",
			tgMon: true, tgDown: false, maxMon: true, maxDown: true,
			wantEmail: true, wantTG: true,
		},
		{
			name: "both down email only",
			tgMon: true, tgDown: true, maxMon: true, maxDown: true,
			wantEmail: true,
		},
		{
			name: "telegram down max disabled email only",
			tgMon: true, tgDown: true, maxMon: false, maxDown: false,
			wantEmail: true,
		},
		{
			name: "max down telegram disabled email only",
			tgMon: false, tgDown: false, maxMon: true, maxDown: true,
			wantEmail: true,
		},
		{
			name: "telegram down max also down email only",
			tgMon: true, tgDown: true, maxMon: true, maxDown: true,
			wantEmail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := planNotificationHealthAlerts(tt.tgMon, tt.tgDown, tt.maxMon, tt.maxDown)
			if plan.SendEmail != tt.wantEmail {
				t.Fatalf("SendEmail = %v, want %v", plan.SendEmail, tt.wantEmail)
			}
			if plan.SendTelegram != tt.wantTG {
				t.Fatalf("SendTelegram = %v, want %v", plan.SendTelegram, tt.wantTG)
			}
			if plan.SendMax != tt.wantMax {
				t.Fatalf("SendMax = %v, want %v", plan.SendMax, tt.wantMax)
			}
		})
	}
}
