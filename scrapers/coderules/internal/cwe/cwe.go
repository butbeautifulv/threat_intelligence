package cwe

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	neo4jstore "coderules/internal/storage/neo4j"
)

const zipURL = "https://cwe.mitre.org/data/xml/cwec_latest.xml.zip"

// IngestFromMITRE downloads the official CWE catalog zip and upserts CWE nodes (name, description).
func IngestFromMITRE(ctx context.Context, st *neo4jstore.Store, maxWeakness int) error {
	if maxWeakness <= 0 {
		maxWeakness = 5000
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, zipURL, nil)
	if err != nil {
		return err
	}
	cl := &http.Client{Timeout: 10 * time.Minute}
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("cwe zip: %s: %s", resp.Status, string(b))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 40<<20))
	if err != nil {
		return err
	}
	br := bytes.NewReader(body)
	zr, err := zip.NewReader(br, int64(len(body)))
	if err != nil {
		return err
	}
	var xmlRC io.ReadCloser
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, ".xml") && strings.HasPrefix(f.Name, "cwec_v") {
			xmlRC, err = f.Open()
			if err != nil {
				return err
			}
			break
		}
	}
	if xmlRC == nil {
		return fmt.Errorf("cwe zip: no cwec_v*.xml found")
	}
	defer xmlRC.Close()

	dec := xml.NewDecoder(xmlRC)
	dec.Strict = false
	n := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "Weakness" {
			continue
		}
		var w struct {
			ID          string `xml:"ID,attr"`
			Name        string `xml:"Name,attr"`
			Status      string `xml:"Status,attr"`
			Description string `xml:"Description"`
		}
		if err := dec.DecodeElement(&w, &se); err != nil {
			continue
		}
		cid := "CWE-" + strings.TrimPrefix(strings.TrimSpace(w.ID), "CWE-")
		desc := strings.TrimSpace(w.Description)
		if err := st.UpsertCWECatalog(ctx, cid, w.Name, desc, w.Status); err != nil {
			return err
		}
		n++
		if n >= maxWeakness {
			break
		}
	}
	return nil
}
