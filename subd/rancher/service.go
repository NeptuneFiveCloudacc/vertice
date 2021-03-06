package rancher

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	nsq "github.com/crackcomm/nsqueue/consumer"
	"github.com/virtengine/libgo/cmd"
	constants "github.com/virtengine/libgo/utils"
	"github.com/virtengine/vertice/carton"
	"github.com/virtengine/vertice/meta"
	"github.com/virtengine/vertice/provision"
)

const (
	TOPIC       = "containers"
	maxInFlight = 150
)

// Service manages the listener and handler for an HTTP endpoint.
type Service struct {
	wg       sync.WaitGroup
	err      chan error
	Handler  *Handler
	Consumer *nsq.Consumer
	Meta     *meta.Config
	Rancherd *Config
}

// NewService returns a new instance of Service.
func NewService(c *meta.Config, d *Config) *Service {
	s := &Service{
		err:      make(chan error),
		Meta:     c,
		Rancherd: d,
	}
	s.Handler = NewHandler(s.Rancherd)
	return s
}

// Open starts the service
func (s *Service) Open() error {
	go func() error {
		log.Info("starting rancherd service")
		if err := nsq.Register(TOPIC, "engine", maxInFlight, s.processNSQ); err != nil {
			return err
		}
		if err := nsq.Connect(s.Meta.NSQd...); err != nil {
			return err
		}
		s.Consumer = nsq.DefaultConsumer

		nsq.Start(true)
		return nil
	}()
	if err := s.setProvisioner(constants.PROVIDER_RANCHER); err != nil {
		return err
	}
	return nil
}

func (s *Service) processNSQ(msg *nsq.Message) {
	log.Debugf(TOPIC + "queue received message  :" + string(msg.Body))
	p, err := carton.NewPayload(msg.Body)
	if err != nil {
		return
	}

	re, err := p.Convert()
	if err != nil {
		return
	}
	go s.Handler.serveNSQ(re)
	return
}

// Close closes the underlying subscribe channel.
func (s *Service) Close() error {
	if s.Consumer != nil {
		s.Consumer.Stop()
	}

	s.wg.Wait()
	return nil
}

// Err returns a channel for fatal errors that occur on the listener.
func (s *Service) Err() <-chan error { return s.err }

//this is an array, a property provider helps to load the provider specific stuff
func (s *Service) setProvisioner(pt string) error {
	var err error
	var tempProv provision.Provisioner

	if tempProv, err = provision.Get(pt); err != nil {
		return err
	}
	log.Debugf(cmd.Colorfy("  > configuring ", "blue", "", "bold") + fmt.Sprintf("%s ", pt))
	if initializableProvisioner, ok := tempProv.(provision.InitializableProvisioner); ok {
		err = initializableProvisioner.Initialize(s.Rancherd.toInterface())
		if err != nil {
			return fmt.Errorf("unable to initialize %s provisioner\n --> %s", pt, err)
		} else {
			log.Debugf(cmd.Colorfy(fmt.Sprintf("  > %s initialized", pt), "blue", "", "bold"))
		}
	}

	if messageProvisioner, ok := tempProv.(provision.MessageProvisioner); ok {
		startupMessage, err := messageProvisioner.StartupMessage()
		if err == nil && startupMessage != "" {
			log.Infof(startupMessage)
		}
	}

	carton.ProvisionerMap[pt] = tempProv
	return nil
}
