package kjv

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCreateDonationCheckoutCreatesOneTimeCheckoutSession(t *testing.T) {
	stripe := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != stripeCheckoutSessionsPath || r.Method != http.MethodPost {
			t.Fatalf("unexpected Stripe request: %s %s", r.Method, r.URL.Path)
		}
		key, _, ok := r.BasicAuth()
		if !ok || key != "sk_test_example" {
			t.Fatalf("missing or incorrect Stripe authentication")
		}
		body, _ := io.ReadAll(r.Body)
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		if form.Get("mode") != "payment" || form.Get("line_items[0][price_data][unit_amount]") != "2500" || form.Get("invoice_creation[enabled]") != "true" {
			t.Fatalf("unexpected checkout form: %v", form)
		}
		if form.Get("success_url") != "https://prsmusa.com/donate/success?session_id={CHECKOUT_SESSION_ID}" {
			t.Fatalf("unexpected success URL: %q", form.Get("success_url"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"url":"https://checkout.stripe.com/c/pay_test"}`))
	}))
	defer stripe.Close()

	app := &App{StripeSecretKey: "sk_test_example", StripeAPIBaseURL: stripe.URL, StripeHTTP: stripe.Client(), PublicBaseURL: "https://prsmusa.com"}
	req := httptest.NewRequest(http.MethodPost, "/donations/checkout", strings.NewReader(`{"amount_cents":2500,"monthly":false}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.createDonationCheckout(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "checkout.stripe.com") {
		t.Fatalf("unexpected response: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCreateDonationCheckoutCreatesMonthlySubscription(t *testing.T) {
	stripe := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("mode") != "subscription" || form.Get("line_items[0][price_data][recurring][interval]") != "month" {
			t.Fatalf("unexpected subscription checkout form: %v", form)
		}
		_, _ = w.Write([]byte(`{"url":"https://checkout.stripe.com/c/sub_test"}`))
	}))
	defer stripe.Close()
	app := &App{StripeSecretKey: "sk_test_example", StripeAPIBaseURL: stripe.URL, StripeHTTP: stripe.Client()}
	req := httptest.NewRequest(http.MethodPost, "/donations/checkout", strings.NewReader(`{"amount_cents":1000,"monthly":true}`))
	rec := httptest.NewRecorder()
	app.createDonationCheckout(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected response: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCreateDonationCheckoutRejectsUnconfiguredOrInvalidDonation(t *testing.T) {
	for _, tt := range []struct {
		name string
		app  *App
		body string
		want int
	}{
		{"unconfigured", &App{}, `{"amount_cents":500}`, http.StatusServiceUnavailable},
		{"too small", &App{StripeSecretKey: "sk_test_example"}, `{"amount_cents":99}`, http.StatusBadRequest},
		{"invalid JSON", &App{StripeSecretKey: "sk_test_example"}, `{`, http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/donations/checkout", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			tt.app.createDonationCheckout(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d; %s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}
