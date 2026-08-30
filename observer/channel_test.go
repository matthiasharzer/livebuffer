package observer_test

import (
	"testing"

	"github.com/matthiasharzer/livebuffer/observer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestSubscriber struct {
	Notifications []string
}

func (s *TestSubscriber) Update(notification string) {
	s.Notifications = append(s.Notifications, notification)
}

func TestNewChannel(t *testing.T) {
	t.Run("sends notifications to a subscriber", func(t *testing.T) {
		testChannel := observer.NewChannel[string]()
		testSubscriber := &TestSubscriber{}
		testChannel.Subscribe(testSubscriber)

		testChannel.Publish("Test Notification 1")
		testChannel.Publish("Test Notification 2")

		require.Len(t, testSubscriber.Notifications, 2)
		assert.Equal(t, "Test Notification 1", testSubscriber.Notifications[0])
		assert.Equal(t, "Test Notification 2", testSubscriber.Notifications[1])
	})

	t.Run("does not send notifications to unsubscribed subscribers", func(t *testing.T) {
		testChannel := observer.NewChannel[string]()
		testSubscriber := &TestSubscriber{}
		testChannel.Subscribe(testSubscriber)
		testChannel.Unsubscribe(testSubscriber)

		testChannel.Publish("Test Notification 1")

		assert.Len(t, testSubscriber.Notifications, 0)
	})

	t.Run("handles multiple subscribers", func(t *testing.T) {
		testChannel := observer.NewChannel[string]()
		subscriber1 := &TestSubscriber{}
		subscriber2 := &TestSubscriber{}
		testChannel.Subscribe(subscriber1)
		testChannel.Subscribe(subscriber2)

		testChannel.Publish("Test Notification")

		require.Len(t, subscriber1.Notifications, 1)
		require.Len(t, subscriber2.Notifications, 1)
		assert.Equal(t, "Test Notification", subscriber1.Notifications[0])
		assert.Equal(t, "Test Notification", subscriber2.Notifications[0])
	})

	t.Run("clears all subscribers", func(t *testing.T) {
		testChannel := observer.NewChannel[string]()
		subscriber1 := &TestSubscriber{}
		subscriber2 := &TestSubscriber{}
		testChannel.Subscribe(subscriber1)
		testChannel.Subscribe(subscriber2)

		testChannel.Clear()

		testChannel.Publish("Test Notification")

		assert.Len(t, subscriber1.Notifications, 0)
		assert.Len(t, subscriber2.Notifications, 0)
	})
}
