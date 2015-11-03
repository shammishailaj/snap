/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PulsePath = os.Getenv("PULSE_PATH")
)

type MockMetricType struct {
	namespace []string
}

func (m MockMetricType) Namespace() []string {
	return m.namespace
}

func (m MockMetricType) LastAdvertisedTime() time.Time {
	return time.Now()
}

func (m MockMetricType) Version() int {
	return 1
}

func (m MockMetricType) Config() *cdata.ConfigDataNode {
	return nil
}

func (m MockMetricType) Data() interface{} {
	return nil
}

func TestCollectPublishWorkflow(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey("Given a started plugin control", t, func() {

		c := control.New()
		c.Start()
		s := New()
		s.SetMetricManager(c)
		Convey("create a workflow", func() {
			rp, _ := core.NewRequestedPlugin(path.Join(PulsePath, "plugin", "pulse-collector-mock2"))
			_, err := c.Load(rp)
			So(err, ShouldBeNil)
			rp2, _ := core.NewRequestedPlugin(path.Join(PulsePath, "plugin", "pulse-publisher-file"))
			_, err = c.Load(rp2)
			So(err, ShouldBeNil)
			rp3, _ := core.NewRequestedPlugin(path.Join(PulsePath, "plugin", "pulse-processor-passthru"))
			_, err = c.Load(rp3)
			So(err, ShouldBeNil)
			time.Sleep(100 * time.Millisecond)

			metrics, err2 := c.MetricCatalog()
			So(err2, ShouldBeNil)
			So(metrics, ShouldNotBeEmpty)

			w := wmap.NewWorkflowMap()
			w.CollectNode.AddMetric("/intel/mock/foo", 2)
			w.CollectNode.AddConfigItem("/intel/mock/foo", "password", "secret")

			pu := wmap.NewPublishNode("file", 3)
			pu.AddConfigItem("file", "/tmp/pulse-TestCollectPublishWorkflow.out")

			pr := wmap.NewProcessNode("passthru", 1)
			time.Sleep(100 * time.Millisecond)

			pr.Add(pu)
			w.CollectNode.Add(pr)

			Convey("Start scheduler", func() {
				err := s.Start()
				So(err, ShouldBeNil)
				Convey("Create task", func() {
					t, err := s.CreateTask(schedule.NewSimpleSchedule(time.Millisecond*500), w, false)
					So(err.Errors(), ShouldBeEmpty)
					So(t, ShouldNotBeNil)
					t.(*task).Spin()
					time.Sleep(3 * time.Second)

				})
			})
		})
	})
}
