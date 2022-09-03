package retail

import (
	"encoding/json"
	"fmt"
	"golang.org/x/xerrors"
	"io"
	"net/http"
	"net/url"
	"time"
)

type svcCaller interface {
	Call(req map[string]interface{}) (io.ReadCloser, error)
}

type PriceCalculator struct {
	priceSvc svcCaller
	vatSvc   svcCaller
}

func NewPriceCalculator(priceSvcEndpoint, vatSvcEndpoint string) *PriceCalculator {
	return &PriceCalculator{
		priceSvc: restEndpointCaller(priceSvcEndpoint),
		vatSvc:   restEndpointCaller(vatSvcEndpoint),
	}
}

func (pc *PriceCalculator) PriceForItem(itemUUID string) (float64, error) {
	return pc.PriceForItemAtDate(itemUUID, time.Now())
}

func (pc *PriceCalculator) PriceForItemAtDate(itemUUID string, date time.Time) (float64, error) {
	priceRes := struct {
		Price float64 `json:"price"`
	}{}

	if err := pc.callService(
		pc.priceSvc,
		map[string]interface{}{
			"item":   itemUUID,
			"period": date,
		},
		&priceRes,
	); err != nil {
		return 0, xerrors.Errorf("unable to retrieve item price: %w", err)
	}

	vatRes := struct {
		Rate float64 `json:"vat_rate"`
	}{}

	if err := pc.callService(
		pc.vatSvc,
		map[string]interface{}{"period": date},
		&vatRes,
	); err != nil {
		return 0, xerrors.Errorf("unable to retrieve vat percent: %w", err)
	}

	return vatInclusivePrice(priceRes.Price, vatRes.Rate), nil
}

func vatInclusivePrice(price, rate float64) float64 {
	return price * (1.0 + rate)
}

func (pc *PriceCalculator) callService(svc svcCaller, req map[string]interface{}, res interface{}) error {
	svcRes, err := svc.Call(req)
	if err != nil {
		return xerrors.Errorf("call to remote service failed: %w", err)
	}
	defer drainAndClose(svcRes)

	if err = json.NewDecoder(svcRes).Decode(res); err != nil {
		return xerrors.Errorf("unable to decode remote service response: %w", err)
	}

	return nil
}

type restEndpointCaller string

func (ep restEndpointCaller) Call(req map[string]interface{}) (io.ReadCloser, error) {
	var params = make(url.Values)
	for k, v := range req {
		params.Set(k, fmt.Sprint(v))
	}
	url := fmt.Sprintf("%s?%s", string(ep), params.Encode())
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		drainAndClose(res.Body)
		return nil, xerrors.Errorf("unexpected response status code: %d", res.StatusCode)
	}

	return res.Body, nil
}

func drainAndClose(r io.ReadCloser) {
	if r == nil {
		return
	}
	io.Copy(io.Discard, r)
	r.Close()
}
