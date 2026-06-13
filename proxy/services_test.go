package proxy

import (
	"errors"
	"testing"
)

func TestResidential_Countries(t *testing.T) {
	d := &fakeDoer{respData: `{
		"continent":{"1":"Asia"},
		"country":[{"continent":"Asia","continent_code":1,
			"list":[{"code":"kr","continent":1,"ip_num":0,"name":"Korea","name_en":"Korea","protocol":1}]}]
	}`}
	res, err := New(d).Residential.Countries(ctx())
	if err != nil {
		t.Fatalf("Countries: %v", err)
	}
	if d.lastPath != "/v1/proxy/domestic_dynamic_area" {
		t.Errorf("path = %q", d.lastPath)
	}
	if res.Continent["1"] != "Asia" || len(res.Country) != 1 || res.Country[0].List[0].Code != "kr" {
		t.Errorf("res = %+v", res)
	}
}

func TestResidential_States(t *testing.T) {
	d := &fakeDoer{respData: `{"data":[{"state":"newyorkcity"}]}`}
	res, err := New(d).Residential.States(ctx(), "us")
	if err != nil {
		t.Fatalf("States: %v", err)
	}
	if d.lastPath != "/v1/proxy/city_by_code" || d.lastFields["code"] != "us" {
		t.Errorf("path=%q fields=%v", d.lastPath, d.lastFields)
	}
	if len(res.Data) != 1 || res.Data[0].State != "newyorkcity" {
		t.Errorf("res = %+v", res)
	}
}

func TestResidential_States_RequiresCode(t *testing.T) {
	_, err := New(&fakeDoer{}).Residential.States(ctx(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, want ValidationError", err)
	}
}

func TestResidential_Cities(t *testing.T) {
	d := &fakeDoer{respData: `{"data":[{"code":"guntersville"}]}`}
	res, err := New(d).Residential.Cities(ctx(), CityParams{Code: "us", Region: "alabama"})
	if err != nil {
		t.Fatalf("Cities: %v", err)
	}
	if d.lastPath != "/v1/proxy/region_by_city" ||
		d.lastFields["code"] != "us" || d.lastFields["region"] != "alabama" {
		t.Errorf("path=%q fields=%v", d.lastPath, d.lastFields)
	}
	if len(res.Data) != 1 || res.Data[0].Code != "guntersville" {
		t.Errorf("res = %+v", res)
	}
}

func TestResidential_Cities_Validation(t *testing.T) {
	_, err := New(&fakeDoer{}).Residential.Cities(ctx(), CityParams{Code: "us"})
	var ve *ValidationError
	if !errors.As(err, &ve) || len(ve.Fields) != 1 || ve.Fields[0] != "region" {
		t.Fatalf("err = %v, want missing region", err)
	}
}

func TestResidential_ISPs(t *testing.T) {
	d := &fakeDoer{respData: `{"list":[{"asn":"AS327712","show_name":"Telecom Algeria"}]}`}
	res, err := New(d).Residential.ISPs(ctx(), "us")
	if err != nil {
		t.Fatalf("ISPs: %v", err)
	}
	if d.lastPath != "/v1/proxy/city_isp" {
		t.Errorf("path = %q", d.lastPath)
	}
	if len(res.List) != 1 || res.List[0].ASN != "AS327712" {
		t.Errorf("res = %+v", res)
	}
}

func TestResidential_Balance(t *testing.T) {
	d := &fakeDoer{respData: `{"balance":124000000111,"expire_time":1769135681}`}
	res, err := New(d).Residential.Balance(ctx())
	if err != nil {
		t.Fatalf("Balance: %v", err)
	}
	if d.lastPath != "/v1/residential_flow/balance" {
		t.Errorf("path = %q", d.lastPath)
	}
	if res.Balance != 124000000111 || res.ExpireTime != 1769135681 {
		t.Errorf("res = %+v", res)
	}
}

func TestResidential_ConsumeLog(t *testing.T) {
	d := &fakeDoer{respData: `{"list":[{"id":1,"uid":81,"balance":30999907705,"all_buy":31000000000,"use":92295,"day":1731427200}]}`}
	res, err := New(d).Residential.ConsumeLog(ctx(), TimeRange{
		Start: "2025-01-01 00:00:00", End: "2025-01-31 23:59:59",
	})
	if err != nil {
		t.Fatalf("ConsumeLog: %v", err)
	}
	if d.lastPath != "/v1/residential_flow/consume_log" ||
		d.lastFields["start_time"] != "2025-01-01 00:00:00" {
		t.Errorf("path=%q fields=%v", d.lastPath, d.lastFields)
	}
	if len(res.List) != 1 || res.List[0].Use != 92295 {
		t.Errorf("res = %+v", res)
	}
}

func TestResidential_ConsumeLog_Validation(t *testing.T) {
	_, err := New(&fakeDoer{}).Residential.ConsumeLog(ctx(), TimeRange{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, want ValidationError", err)
	}
	assertStringSet(t, ve.Fields, []string{"start_time", "end_time"})
}

func TestMobile_All(t *testing.T) {
	// Countries returns raw payload.
	dc := &fakeDoer{}
	dc.respData = `{"foo":1}`
	raw, err := New(dc).Mobile.Countries(ctx())
	if err != nil {
		t.Fatalf("Countries: %v", err)
	}
	if dc.lastPath != "/v1/proxy/mobile_area" || string(raw) != `{"foo":1}` {
		t.Errorf("path=%q raw=%s", dc.lastPath, raw)
	}

	db := &fakeDoer{respData: `{"balance":30999899170}`}
	bal, err := New(db).Mobile.Balance(ctx())
	if err != nil || bal.Balance != 30999899170 {
		t.Fatalf("Balance: %v %+v", err, bal)
	}

	du := &fakeDoer{respData: `{"list":[{"use":92295}]}`}
	use, err := New(du).Mobile.ConsumeLog(ctx(), MobileUseParams{
		Start: "a", End: "b", Granularity: "2",
	})
	if err != nil {
		t.Fatalf("ConsumeLog: %v", err)
	}
	if du.lastFields["day_or_hour"] != "2" || use.List[0].Use != 92295 {
		t.Errorf("fields=%v res=%+v", du.lastFields, use)
	}
}

func TestMobile_ConsumeLog_Validation(t *testing.T) {
	_, err := New(&fakeDoer{}).Mobile.ConsumeLog(ctx(), MobileUseParams{Start: "a"})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v", err)
	}
	assertStringSet(t, ve.Fields, []string{"end_time", "day_or_hour"})
}

func TestRotatingISP_All(t *testing.T) {
	dc := &fakeDoer{respData: `{"country":[{"code":"kw","node":"Kuwait"}]}`}
	c, err := New(dc).RotatingISP.Countries(ctx())
	if err != nil || dc.lastPath != "/v1/proxy/isp_data_area" || c.Country[0].Code != "kw" {
		t.Fatalf("Countries: %v %+v", err, c)
	}
	db := &fakeDoer{respData: `{"balance":1,"expire_time":2}`}
	if _, err := New(db).RotatingISP.Balance(ctx()); err != nil || db.lastPath != "/v1/isp_flow/balance" {
		t.Fatalf("Balance: %v path=%q", err, db.lastPath)
	}
	dl := &fakeDoer{respData: `{"list":[]}`}
	if _, err := New(dl).RotatingISP.ConsumeLog(ctx(), TimeRange{Start: "a", End: "b"}); err != nil ||
		dl.lastPath != "/v1/isp_flow/consume_log" {
		t.Fatalf("ConsumeLog: %v path=%q", err, dl.lastPath)
	}
}

func TestRotatingDC_All(t *testing.T) {
	dc := &fakeDoer{respData: `{"country":[{"id":1,"code":"us","city":{"code":"random"}}]}`}
	c, err := New(dc).RotatingDC.Countries(ctx())
	if err != nil || dc.lastPath != "/v1/proxy/dynamic_data_area" || c.Country[0].City.Code != "random" {
		t.Fatalf("Countries: %v %+v", err, c)
	}
	db := &fakeDoer{respData: `{"balance":1}`}
	if _, err := New(db).RotatingDC.Balance(ctx()); err != nil || db.lastPath != "/v1/dc_flow/balance" {
		t.Fatalf("Balance: %v", err)
	}
	dl := &fakeDoer{respData: `{"list":[]}`}
	if _, err := New(dl).RotatingDC.ConsumeLog(ctx(), TimeRange{Start: "a", End: "b"}); err != nil ||
		dl.lastPath != "/v1/dc_flow/consume_log" {
		t.Fatalf("ConsumeLog: %v", err)
	}
}

func TestStaticISP_Open(t *testing.T) {
	d := &fakeDoer{}
	err := New(d).StaticISP.Open(ctx(), OpenStaticISPParams{
		IPType: "normal", Region: "hk:1|us-va:2", Duration: "week", Num: 3,
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	assertFields(t, d.lastFields, map[string]string{
		"ip_type": "normal", "region": "hk:1|us-va:2", "duration": "week", "num": "3",
	})
	if d.lastPath != "/v1/static_house/open" {
		t.Errorf("path = %q", d.lastPath)
	}
}

func TestStaticISP_Open_Validation(t *testing.T) {
	err := New(&fakeDoer{}).StaticISP.Open(ctx(), OpenStaticISPParams{Region: "hk:1"})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v", err)
	}
	assertStringSet(t, ve.Fields, []string{"ip_type", "duration", "num"})
}

func TestStaticISP_List_DefaultsAndFilters(t *testing.T) {
	d := &fakeDoer{respData: `{"list":[{"id":51,"ip":"156.249.22.144","status":2,"auto_renew":0}],"page":1,"total":13}`}
	auto := 1
	res, err := New(d).StaticISP.List(ctx(), ListStaticParams{
		Status: "2", Region: "us-va", IsAutoRenew: &auto,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if d.lastFields["page"] != "1" || d.lastFields["limit"] != "10" ||
		d.lastFields["status"] != "2" || d.lastFields["is_auto_renew"] != "1" {
		t.Errorf("fields = %v", d.lastFields)
	}
	if res.Total != 13 || res.List[0].IP != "156.249.22.144" {
		t.Errorf("res = %+v", res)
	}
}

func TestStaticISP_Export(t *testing.T) {
	d := &fakeDoer{rawResp: []byte("id,ip\n1,1.2.3.4\n")}
	b, err := New(d).StaticISP.Export(ctx(), ExportStaticParams{Status: "1"})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if d.lastPath != "/v1/static_house/export" || string(b) != "id,ip\n1,1.2.3.4\n" {
		t.Errorf("path=%q body=%q", d.lastPath, b)
	}
}

func TestStaticISP_Region(t *testing.T) {
	d := &fakeDoer{respData: `{"list":{"Asia-Pacific":[{"node":"Dubai","node_en":"Dubai","param":0,"region":"uae-dubai"}]}}`}
	res, err := New(d).StaticISP.Region(ctx(), "isp-resi-hq")
	if err != nil {
		t.Fatalf("Region: %v", err)
	}
	if d.lastFields["isp_type"] != "isp-resi-hq" {
		t.Errorf("fields = %v", d.lastFields)
	}
	if res.List["Asia-Pacific"][0].Region != "uae-dubai" {
		t.Errorf("res = %+v", res)
	}
}

func TestStaticISP_Region_RequiresType(t *testing.T) {
	_, err := New(&fakeDoer{}).StaticISP.Region(ctx(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v", err)
	}
}

func TestStaticISP_Renew(t *testing.T) {
	d := &fakeDoer{}
	if err := New(d).StaticISP.Renew(ctx(), RenewStaticParams{
		IPs: "127.0.0.1,127.0.0.2", Duration: "week",
	}); err != nil {
		t.Fatalf("Renew: %v", err)
	}
	assertFields(t, d.lastFields, map[string]string{
		"renew_ip_list": "127.0.0.1,127.0.0.2", "duration": "week",
	})
}

func TestStaticISP_RenewSetting_InjectsType(t *testing.T) {
	d := &fakeDoer{}
	if err := New(d).StaticISP.RenewSetting(ctx(), RenewSettingParams{
		IDs: "100,111", PackageType: "week", Status: 1, RenewType: 1,
	}); err != nil {
		t.Fatalf("RenewSetting: %v", err)
	}
	if d.lastFields["type"] != "static_house" {
		t.Errorf("type field = %q, want static_house", d.lastFields["type"])
	}
	assertFields(t, d.lastFields, map[string]string{
		"type": "static_house", "ids": "100,111", "package_type": "week",
		"status": "1", "renew_type": "1",
	})
}

func TestStaticISP_RenewSetting_Validation(t *testing.T) {
	err := New(&fakeDoer{}).StaticISP.RenewSetting(ctx(), RenewSettingParams{IDs: "1"})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v", err)
	}
	assertStringSet(t, ve.Fields, []string{"package_type", "status", "renew_type"})
}

func TestDedicatedDC_OpenAndPaths(t *testing.T) {
	d := &fakeDoer{}
	if err := New(d).DedicatedDC.Open(ctx(), OpenDedicatedDCParams{
		Region: "hk:1|tw:2", Duration: "week", Num: 3,
	}); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if d.lastPath != "/v1/static/open" {
		t.Errorf("path = %q", d.lastPath)
	}
	if _, ok := d.lastFields["ip_type"]; ok {
		t.Errorf("dedicated DC open must not send ip_type: %v", d.lastFields)
	}

	dr := &fakeDoer{respData: `{"list":{"Oceania":[{"node":"Sydney","region":"au-syd"}]}}`}
	reg, err := New(dr).DedicatedDC.Region(ctx())
	if err != nil || dr.lastPath != "/v1/static/region" || reg.List["Oceania"][0].Region != "au-syd" {
		t.Fatalf("Region: %v %+v", err, reg)
	}

	drs := &fakeDoer{}
	if err := New(drs).DedicatedDC.RenewSetting(ctx(), RenewSettingParams{
		IDs: "1", PackageType: "week", Status: 1, RenewType: 2,
	}); err != nil {
		t.Fatalf("RenewSetting: %v", err)
	}
	if drs.lastFields["type"] != "static" {
		t.Errorf("type = %q, want static", drs.lastFields["type"])
	}
}

func TestUnlimited_Hosts(t *testing.T) {
	d := &fakeDoer{respData: `{"list":[{"id":277,"host":"yyds.com","band_width":0,"hardware":"16_32"}],"page":1,"total":1}`}
	res, err := New(d).Unlimited.Hosts(ctx(), 0, 0) // defaults
	if err != nil {
		t.Fatalf("Hosts: %v", err)
	}
	if d.lastPath != "/v1/unlimited/host_list" ||
		d.lastFields["page"] != "1" || d.lastFields["limit"] != "10" {
		t.Errorf("path=%q fields=%v", d.lastPath, d.lastFields)
	}
	if res.Total != 1 || res.List[0].Host != "yyds.com" {
		t.Errorf("res = %+v", res)
	}
}

func TestProhibitDomain_AddListDelete(t *testing.T) {
	da := &fakeDoer{}
	if err := New(da).ProhibitDomain.Add(ctx(), "www.baidu.com"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if da.lastPath != "/v1/prohibit_domain/add" || da.lastFields["address"] != "www.baidu.com" {
		t.Errorf("path=%q fields=%v", da.lastPath, da.lastFields)
	}

	if err := New(&fakeDoer{}).ProhibitDomain.Add(ctx(), ""); err == nil {
		t.Error("Add with empty address should fail validation")
	}

	dl := &fakeDoer{respData: `{"list":[{"id":231,"address":"google.com","status":1}],"total":1}`}
	res, err := New(dl).ProhibitDomain.List(ctx())
	if err != nil || res.Total != 1 || res.List[0].Address != "google.com" {
		t.Fatalf("List: %v %+v", err, res)
	}

	// Delete by ID.
	dd := &fakeDoer{}
	if err := New(dd).ProhibitDomain.Delete(ctx(), DeleteProhibitParams{ID: "231"}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if dd.lastFields["is_all"] != "2" || dd.lastFields["id"] != "231" {
		t.Errorf("fields = %v", dd.lastFields)
	}

	// Delete all.
	dall := &fakeDoer{}
	if err := New(dall).ProhibitDomain.Delete(ctx(), DeleteProhibitParams{All: true}); err != nil {
		t.Fatalf("Delete all: %v", err)
	}
	if dall.lastFields["is_all"] != "1" {
		t.Errorf("is_all = %q, want 1", dall.lastFields["is_all"])
	}

	// Delete without ID and not All -> validation error.
	if err := New(&fakeDoer{}).ProhibitDomain.Delete(ctx(), DeleteProhibitParams{}); err == nil {
		t.Error("Delete without id or All should fail validation")
	}
}
