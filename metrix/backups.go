package metrix

import (
	"time"

	constants "github.com/virtengine/libgo/utils"
	"github.com/virtengine/vertice/carton"
)

const (
	BACKUPS        = "backups"
	BACKUPS_SENSOR = "instance.backups.exists"
)

type Backups struct {
	DefaultUnits map[string]string
	RawStatus    []byte
}

func (r *Backups) Prefix() string {
	return BACKUPS
}

func (r *Backups) DeductBill(c *MetricsCollection) (e error) {
	for _, mc := range c.Sensors {
		mkBalance(mc, r.DefaultUnits)
	}
	return
}

func (s *Backups) Collect(c *MetricsCollection) (e error) {
	bk := carton.Backups{}
	bks, e := bk.GetBox()
	if e != nil {
		return
	}

	s.CollectMetricsFromStats(c, bks)
	e = s.DeductBill(c)
	return
}

func (c *Backups) ReadUsers() ([]*carton.Account, error) {
	act := new(carton.Account)
	res, e := act.GetUsers()
	if e != nil {
		return nil, e
	}
	return res, nil
}

//actually the NewSensor can create trypes based on the event type.
func (c *Backups) CollectMetricsFromStats(mc *MetricsCollection, bks []carton.Backups) {
	for _, a := range bks {
		sc := NewSensor(BACKUPS_SENSOR)
		sc.AccountId = a.AccountId
		sc.AssemblyId = a.AssemblyId
		sc.System = c.Prefix()
		sc.Node = ""
		sc.AssemblyName = a.Name
		sc.AssembliesId = a.Id
		sc.Source = c.Prefix()
		sc.Message = "backups billing"
		sc.Status = "health-ok"
		sc.AuditPeriodBeginning = time.Now().Add(-MetricsInterval).Format(time.RFC3339)
		sc.AuditPeriodEnding = time.Now().Format(time.RFC3339)
		sc.AuditPeriodDelta = ""
		sc.addMetric(STORAGE_COST, c.DefaultUnits[STORAGE_COST_PER_HOUR], a.Sizeof(), "delta")
		sc.CreatedAt = time.Now()
		if a.Status == constants.IMAGE_READY {
			mc.Add(sc)
		}
	}

	return
}
