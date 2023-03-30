// Copyright 2021-2023 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cerbos/cerbos/internal/audit"
	"github.com/twmb/franz-go/pkg/kgo"
)

const confKey = audit.ConfKey + ".kafka"

const (
	defaultAcknowledgement    = AckAll
	defaultEncoding           = EncodingJSON
	defaultFlushTimeout       = 30 * time.Second
	defaultClientID           = "cerbos"
	defaultMaxBufferedRecords = 250
)

// Conf is optional configuration for kafka Audit.
type Conf struct {
	// Ack mode for producing messages. Valid values are "none", "leader" or "all" (default). Idempotency is disabled when mode is not "all".
	Ack string `yaml:"ack" conf:",example=all"`
	// Topic to write audit entries to.
	Topic string `yaml:"topic" conf:"required,example=cerbos.audit.log"`
	// Encoding format. Valid values are "json" (default) or "protobuf".
	Encoding Encoding `yaml:"encoding" conf:",example=json"`
	// ClientID reported in Kafka connections.
	ClientID string `yaml:"clientID" conf:",example=cerbos"`
	// Brokers list to seed the Kafka client.
	Brokers []string `yaml:"brokers" conf:"required,example=['localhost:9092']"`
	// FlushTimeout sets how often messages are flushed to the remote Kafka server.
	FlushTimeout time.Duration `yaml:"flushTimeout" conf:",example=30s"`
	// MaxBufferedRecords sets the maximum number of records the client should buffer in memory in async mode.
	MaxBufferedRecords int `yaml:"maxBufferedRecords" conf:",example=1000"`
	// ProduceSync forces the client to produce messages to Kafka synchronously. This can have a significant impact on performance.
	ProduceSync bool `yaml:"produceSync" conf:",example=false"`
}

func (c *Conf) Key() string {
	return confKey
}

func (c *Conf) SetDefaults() {
	c.Ack = defaultAcknowledgement
	c.Encoding = defaultEncoding
	c.FlushTimeout = defaultFlushTimeout
	c.ClientID = defaultClientID
	c.MaxBufferedRecords = defaultMaxBufferedRecords
}

func (c *Conf) Validate() error {
	if _, err := formatAck(c.Ack); err != nil {
		return err
	}

	if strings.TrimSpace(c.Topic) == "" {
		return errors.New("invalid topic")
	}

	switch c.Encoding {
	case EncodingJSON, EncodingProtobuf:
	default:
		return fmt.Errorf("invalid encoding format: %s", c.Encoding)
	}

	if c.FlushTimeout <= 0 {
		return errors.New("invalid flush timeout")
	}

	if strings.TrimSpace(c.ClientID) == "" {
		return errors.New("invalid client ID")
	}

	if len(c.Brokers) == 0 {
		return errors.New("empty brokers")
	}

	return nil
}

func formatAck(ack string) (kgo.Acks, error) {
	switch ack {
	case AckNone:
		return kgo.NoAck(), nil
	case AckAll:
		return kgo.AllISRAcks(), nil
	case AckLeader:
		return kgo.LeaderAck(), nil
	default:
		return kgo.NoAck(), fmt.Errorf("invalid ack value: %s", ack)
	}
}