package reconciliation

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/reconciler/pkg/db"
	"github.com/kyma-incubator/reconciler/pkg/model"
)

type FilterMixer struct {
	Filters []Filter
}

func (fm *FilterMixer) FilterByQuery(q *db.Select) error {
	for i := range fm.Filters {
		if err := fm.Filters[i].FilterByQuery(q); err != nil {
			return err
		}
	}
	return nil
}

func (fm *FilterMixer) FilterByInstance(re *model.ReconciliationEntity) *model.ReconciliationEntity {
	entity := re
	for i := range fm.Filters {
		entity = fm.Filters[i].FilterByInstance(entity)
		if entity == nil {
			break
		}
	}
	return entity
}

type Limit struct {
	Count       int
	actualCount int
}

func (l *Limit) FilterByQuery(q *db.Select) error {
	q.OrderBy(map[string]string{"Created": "DESC"}).Limit(l.Count)
	return nil
}

func (l *Limit) FilterByInstance(re *model.ReconciliationEntity) *model.ReconciliationEntity {
	if l.actualCount < l.Count {
		l.actualCount++
		return re
	}
	return nil
}

type WithStatuses struct {
	Statuses []string
}

func (ws *WithStatuses) FilterByQuery(q *db.Select) error {
	if len(ws.Statuses) < 1 {
		return nil
	}

	column, err := columnName(q, "Status")
	if err != nil {
		return err
	}

	var whereRaw string
	argsOffset := q.NextPlaceholderCount()
	for i := range ws.Statuses {
		if i > 0 {
			whereRaw = fmt.Sprintf("%s%s", whereRaw, " OR ")
		}
		whereRaw = fmt.Sprintf("%s%s=$%d", whereRaw, column, argsOffset+i)
	}

	q.WhereRaw(whereRaw, toInterfaceSlice(ws.Statuses)...)

	return nil
}

func (ws *WithStatuses) FilterByInstance(re *model.ReconciliationEntity) *model.ReconciliationEntity {
	if len(ws.Statuses) < 1 {
		return nil
	}

	for i := range ws.Statuses {
		if ws.Statuses[i] == string(re.Status) {
			return re
		}
	}
	return nil
}

type WithCreationDateAfter struct {
	Time time.Time
}

func (wd *WithCreationDateAfter) FilterByQuery(q *db.Select) error {
	column, err := columnName(q, "Created")
	if err != nil {
		return err
	}

	q.WhereRaw(fmt.Sprintf("%s>$%d", column, q.NextPlaceholderCount()), wd.Time.Format("2006-01-02 15:04:05.000"))
	return nil
}

func (wd *WithCreationDateAfter) FilterByInstance(i *model.ReconciliationEntity) *model.ReconciliationEntity {
	if i.Created.After(wd.Time) {
		return i
	}
	return nil
}

type WithCreationDateBefore struct {
	Time time.Time
}

func (wd *WithCreationDateBefore) FilterByQuery(q *db.Select) error {
	column, err := columnName(q, "Created")
	if err != nil {
		return err
	}

	q.WhereRaw(fmt.Sprintf("%s<$%d", column, q.NextPlaceholderCount()), wd.Time.Format("2006-01-02 15:04:05.000"))
	return nil
}

func (wd *WithCreationDateBefore) FilterByInstance(i *model.ReconciliationEntity) *model.ReconciliationEntity {
	if i.Created.Before(wd.Time) {
		return i
	}
	return nil
}

type WithSchedulingID struct {
	SchedulingID string
}

func (ws *WithSchedulingID) FilterByQuery(q *db.Select) error {
	q.Where(map[string]interface{}{
		"SchedulingID": ws.SchedulingID,
	})
	return nil
}

func (ws *WithSchedulingID) FilterByInstance(i *model.ReconciliationEntity) *model.ReconciliationEntity {
	if i.SchedulingID == ws.SchedulingID {
		return i
	}
	return nil
}

type WithRuntimeIDs struct {
	RuntimeIDs []string
}

func (wc *WithRuntimeIDs) FilterByQuery(q *db.Select) error {
	runtimeIDsLen := len(wc.RuntimeIDs)
	if runtimeIDsLen < 1 {
		return nil
	}

	var values string
	argsOffset := q.NextPlaceholderCount()
	for i := range wc.RuntimeIDs {
		if i == runtimeIDsLen-1 {
			values = fmt.Sprintf("%s$%d", values, i+argsOffset)
			break
		}
		values = fmt.Sprintf("%s$%d,", values, i+argsOffset)
	}

	runtimeIDs := toInterfaceSlice(wc.RuntimeIDs)
	q.WhereIn("RuntimeID", values, runtimeIDs...)
	return nil
}

func (wc *WithRuntimeIDs) FilterByInstance(i *model.ReconciliationEntity) *model.ReconciliationEntity {
	runtimeIDsLen := len(wc.RuntimeIDs)
	if runtimeIDsLen < 1 {
		return nil
	}

	for _, runtimeID := range wc.RuntimeIDs {
		if i.RuntimeID == runtimeID {
			return i
		}
	}
	return nil
}

type WithRuntimeID struct {
	RuntimeID string
}

func (wc *WithRuntimeID) FilterByQuery(q *db.Select) error {
	q.Where(map[string]interface{}{
		"RuntimeID": wc.RuntimeID,
	})
	return nil
}

func (wc *WithRuntimeID) FilterByInstance(i *model.ReconciliationEntity) *model.ReconciliationEntity {
	if i.RuntimeID == wc.RuntimeID {
		return i
	}
	return nil
}

type CurrentlyReconciling struct {
}

func (cr *CurrentlyReconciling) FilterByQuery(q *db.Select) error {
	q.Where(map[string]interface{}{
		"Finished": false,
	})
	return nil
}

func (cr *CurrentlyReconciling) FilterByInstance(i *model.ReconciliationEntity) *model.ReconciliationEntity {
	if !i.Finished {
		return i
	}
	return nil
}

type CurrentlyReconcilingWithRuntimeID struct {
	RuntimeID string
}

func (cr *CurrentlyReconcilingWithRuntimeID) FilterByQuery(q *db.Select) error {
	q.Where(map[string]interface{}{
		"Finished":  false,
		"RuntimeID": cr.RuntimeID,
	})
	return nil
}

func (cr *CurrentlyReconcilingWithRuntimeID) FilterByInstance(i *model.ReconciliationEntity) *model.ReconciliationEntity {
	if !i.Finished && i.RuntimeID == cr.RuntimeID {
		return i
	}
	return nil
}

func toInterfaceSlice(args []string) []interface{} {
	argsLen := len(args)
	result := make([]interface{}, argsLen)
	for i := 0; i < argsLen; i++ {
		result[i] = args[i]
	}
	return result
}

type WithClusterConfigStatus struct {
	ClusterConfigStatus int64
}

func (wc *WithClusterConfigStatus) FilterByQuery(q *db.Select) error {
	q.Where(map[string]interface{}{
		"ClusterConfigStatus": wc.ClusterConfigStatus,
	})
	return nil
}

func (wc *WithClusterConfigStatus) FilterByInstance(i *model.ReconciliationEntity) *model.ReconciliationEntity {
	if i.ClusterConfigStatus == wc.ClusterConfigStatus {
		return i
	}
	return nil
}

func columnName(q *db.Select, name string) (string, error) {
	statusColHandler, err := db.NewColumnHandler(&model.ReconciliationEntity{}, q.Conn, q.Logger)
	if err != nil {
		return "", err
	}
	return statusColHandler.ColumnName(name)
}
