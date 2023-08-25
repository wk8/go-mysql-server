// Copyright 2023 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

// loadEvents loads any events that are required for a plan node to operate properly (except for nodes dealing with
// event execution).
func loadEvents(ctx *sql.Context, a *Analyzer, n sql.Node, scope *plan.Scope, sel RuleSelector) (sql.Node, transform.TreeIdentity, error) {
	span, ctx := ctx.Span("loadEvents")
	defer span.End()

	return transform.Node(n, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		switch node := n.(type) {
		case *plan.ShowEvents:
			newShowEvents := *node
			loadedEvents, err := loadEventsFromDb(ctx, a.Catalog, newShowEvents.Database())
			if err != nil {
				return nil, transform.SameTree, err
			}
			newShowEvents.Events = loadedEvents
			return &newShowEvents, transform.NewTree, nil
		case *plan.ShowCreateEvent:
			newShowCreateEvent := *node
			loadedEvent, err := loadEventFromDb(ctx, a.Catalog, newShowCreateEvent.Database(), newShowCreateEvent.EventName)
			if err != nil {
				return nil, transform.SameTree, err
			}
			newShowCreateEvent.Event = loadedEvent
			return &newShowCreateEvent, transform.NewTree, nil
		case *plan.AlterEvent:
			newAlterEvent := *node
			loadedEvent, err := loadEventFromDb(ctx, a.Catalog, newAlterEvent.Database(), newAlterEvent.EventName)
			if err != nil {
				return nil, transform.SameTree, err
			}
			newAlterEvent.Event = loadedEvent
			return &newAlterEvent, transform.NewTree, nil
		default:
			return node, transform.SameTree, nil
		}
	})
}

func loadEventsFromDb(ctx *sql.Context, cat sql.Catalog, db sql.Database) ([]sql.EventDetails, error) {
	var loadedEvents []sql.EventDetails
	if eventDb, ok := db.(sql.EventDatabase); ok {
		events, err := eventDb.GetEvents(ctx)
		if err != nil {
			return nil, err
		}
		for _, event := range events {
			ed, err := GetEventDetailsFromEventDefinition(ctx, cat, event)
			if err != nil {
				return nil, err
			}
			loadedEvents = append(loadedEvents, ed)
		}
	}
	return loadedEvents, nil
}

func loadEventFromDb(ctx *sql.Context, cat sql.Catalog, db sql.Database, name string) (sql.EventDetails, error) {
	eventDb, ok := db.(sql.EventDatabase)
	if !ok {
		return sql.EventDetails{}, sql.ErrEventsNotSupported.New(db.Name())
	}

	event, exists, err := eventDb.GetEvent(ctx, name)
	if err != nil {
		return sql.EventDetails{}, err
	} else if !exists {
		return sql.EventDetails{}, sql.ErrUnknownEvent.New(name)
	}
	return GetEventDetailsFromEventDefinition(ctx, cat, event)
}

// GetEventDetailsFromEventDefinition returns sql.EventDetails by parsing its CREATE statement.
// All time values of events will be evaluated in UTC TZ as they are stored UTC TZ.
func GetEventDetailsFromEventDefinition(ctx *sql.Context, cat sql.Catalog, event sql.EventDefinition) (sql.EventDetails, error) {
	b, err := planbuilder.New(ctx, cat)
	if err != nil {
		return sql.EventDetails{}, err
	}
	b.SetParserOptions(sql.NewSqlModeFromString(event.SqlMode).ParserOptions())
	parsedCreateEvent, _, _, err := b.Parse(event.CreateStatement, false)
	if err != nil {
		return sql.EventDetails{}, err
	}
	eventPlan, ok := parsedCreateEvent.(*plan.CreateEvent)
	if !ok {
		return sql.EventDetails{}, sql.ErrEventCreateStatementInvalid.New(event.CreateStatement)
	}
	return eventPlan.GetEventDetails(ctx, event.CreatedAt, event.LastAltered, event.LastExecuted, "+00:00")
}
