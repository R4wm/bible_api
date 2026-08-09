package kjv

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	stripeCheckoutSessionsPath = "/v1/checkout/sessions"
	minimumDonationCents       = 100
	maximumDonationCents       = 1000000
)

type donationCheckoutRequest struct {
	AmountCents int  `json:"amount_cents"`
	Monthly     bool `json:"monthly"`
}

type stripeCheckoutSession struct {
	URL string `json:"url"`
}

func (app *App) SetupDonationRoutes() {
	app.Router.HandleFunc("/donate", app.donatePage).Methods("GET")
	app.Router.HandleFunc("/donate/success", app.donationSuccessPage).Methods("GET")
	app.Router.HandleFunc("/donations/checkout", app.createDonationCheckout).Methods("POST")
}

func (app *App) donatePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = donationPageTemplate.Execute(w, nil)
}

func (app *App) donationSuccessPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = donationSuccessTemplate.Execute(w, nil)
}

func (app *App) createDonationCheckout(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(app.StripeSecretKey) == "" {
		jsonError(w, http.StatusServiceUnavailable, "donations are not configured")
		return
	}

	var input donationCheckoutRequest
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid donation request")
		return
	}
	if input.AmountCents < minimumDonationCents || input.AmountCents > maximumDonationCents {
		jsonError(w, http.StatusBadRequest, "donation amount must be between $1.00 and $10,000.00")
		return
	}

	checkoutURL, err := app.createStripeDonationCheckout(r, input)
	if err != nil {
		jsonError(w, http.StatusBadGateway, "unable to start secure checkout")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": checkoutURL})
}

func (app *App) createStripeDonationCheckout(r *http.Request, input donationCheckoutRequest) (string, error) {
	baseURL := strings.TrimRight(app.StripeAPIBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.stripe.com"
	}
	publicURL := app.donationPublicURL(r)

	form := url.Values{
		"line_items[0][price_data][currency]":           {"usd"},
		"line_items[0][price_data][product_data][name]": {"Bible API Donation"},
		"line_items[0][price_data][unit_amount]":        {strconv.Itoa(input.AmountCents)},
		"line_items[0][quantity]":                       {"1"},
		"success_url":                                   {publicURL + "/donate/success?session_id={CHECKOUT_SESSION_ID}"},
		"cancel_url":                                    {publicURL + "/donate"},
		"metadata[purpose]":                             {"bible_api_donation"},
	}
	if input.Monthly {
		form.Set("mode", "subscription")
		form.Set("line_items[0][price_data][recurring][interval]", "month")
	} else {
		form.Set("mode", "payment")
		form.Set("invoice_creation[enabled]", "true")
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, baseURL+stripeCheckoutSessionsPath, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(app.StripeSecretKey, "")
	client := app.StripeHTTP
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("stripe returned status %d", resp.StatusCode)
	}

	var session stripeCheckoutSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return "", err
	}
	if session.URL == "" {
		return "", fmt.Errorf("stripe response did not include a checkout URL")
	}
	return session.URL, nil
}

func (app *App) donationPublicURL(r *http.Request) string {
	if app.PublicBaseURL != "" {
		return strings.TrimRight(app.PublicBaseURL, "/")
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

var donationPageTemplate = template.Must(template.New("donate").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Support Bible API</title><style>
body{margin:0;background:#f8f5f1;color:#251c1a;font-family:Arial,sans-serif}.card{max-width:560px;margin:48px auto;padding:32px;background:white;border:1px solid #e5d9d3;border-radius:16px;box-shadow:0 8px 28px #3b24151a}h1{margin-top:0;color:#7a1818}fieldset{border:0;padding:0;margin:24px 0}legend{font-weight:bold;margin-bottom:10px}.choices{display:grid;grid-template-columns:repeat(4,1fr);gap:8px}button{border:1px solid #c93333;border-radius:999px;background:#fffafa;color:#7a1818;padding:11px;cursor:pointer;font-size:16px}button[aria-pressed="true"],button:hover{background:#ffe8e8}.primary{width:100%;background:#7a1818;color:white;margin-top:18px}.primary:hover{background:#571010}input{width:100%;padding:11px;border:1px solid #bba9a0;border-radius:8px;font-size:16px}p{line-height:1.5}.note,#error{font-size:14px;color:#684d44}#error{color:#a41111;min-height:1.3em}a{color:#7a1818}</style></head>
<body><main class="card"><a href="/v2/">← Back to Bible API</a><h1>Support Bible API</h1><p>Your gift helps keep Bible API available, focused, and free of advertising.</p>
<fieldset><legend>Donation frequency</legend><div class="choices"><button type="button" data-monthly="false" aria-pressed="true">One-time</button><button type="button" data-monthly="true" aria-pressed="false">Monthly</button></div></fieldset>
<fieldset><legend>Amount (USD)</legend><div class="choices" id="amounts"><button type="button" data-amount="500" aria-pressed="true">$5</button><button type="button" data-amount="1000">$10</button><button type="button" data-amount="2500">$25</button><button type="button" data-amount="5000">$50</button></div><p><label for="custom">Or enter another amount</label><input id="custom" type="number" min="1" max="10000" step="0.01" inputmode="decimal" placeholder="Amount in USD"></p></fieldset>
<div id="error" role="alert"></div><button class="primary" id="donate" type="button">Continue to secure checkout</button><p class="note">Payments are securely processed by Stripe. You will be redirected to Stripe Checkout to complete your donation.</p></main>
<script>let monthly=false,amount=500;const qs=s=>document.querySelector(s),all=s=>document.querySelectorAll(s);all('[data-monthly]').forEach(b=>b.onclick=()=>{monthly=b.dataset.monthly==='true';all('[data-monthly]').forEach(x=>x.setAttribute('aria-pressed',x===b));});all('[data-amount]').forEach(b=>b.onclick=()=>{amount=Number(b.dataset.amount);qs('#custom').value='';all('[data-amount]').forEach(x=>x.setAttribute('aria-pressed',x===b));});qs('#donate').onclick=async()=>{const custom=qs('#custom').value;if(custom)amount=Math.round(Number(custom)*100);const error=qs('#error');error.textContent='';if(!Number.isInteger(amount)||amount<100||amount>1000000){error.textContent='Enter an amount from $1.00 to $10,000.00.';return}const button=qs('#donate');button.disabled=true;button.textContent='Opening secure checkout…';try{const res=await fetch('/donations/checkout',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({amount_cents:amount,monthly})});const data=await res.json();if(!res.ok)throw new Error(data.error||'Checkout is unavailable.');location.assign(data.url)}catch(e){error.textContent=e.message;button.disabled=false;button.textContent='Continue to secure checkout'}};</script></body></html>`))

var donationSuccessTemplate = template.Must(template.New("donate-success").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Thank You</title><style>body{margin:0;background:#f8f5f1;color:#251c1a;font-family:Arial,sans-serif}.card{max-width:560px;margin:48px auto;padding:32px;background:white;border:1px solid #e5d9d3;border-radius:16px;box-shadow:0 8px 28px #3b24151a}a{display:inline-block;margin-top:12px;color:#7a1818}</style></head><body><main class="card"><h1>Thank you for supporting Bible API.</h1><p>Your donation has been received. Stripe will provide your payment confirmation and any available receipt.</p><a href="/v2/">Return to Bible API</a></main></body></html>`))
