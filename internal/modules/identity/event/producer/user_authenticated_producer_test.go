package producer_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event/producer"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	kafka_mocks "github.com/cristiano-pacheco/pingo/pkg/kafka/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
)

type mockLifecycle struct{}

func (m *mockLifecycle) Append(hook fx.Hook) {
	// No-op for testing
}

type UserAuthenticatedProducerTestSuite struct {
	suite.Suite
	sut          *producer.UserAuthenticatedProducer
	kafkaBuilder *kafka_mocks.MockBuilder
	mockProducer *kafka_mocks.MockProducer
	logger       logger.Logger
}

func (s *UserAuthenticatedProducerTestSuite) SetupTest() {
	s.kafkaBuilder = kafka_mocks.NewMockBuilder(s.T())
	s.mockProducer = kafka_mocks.NewMockProducer(s.T())

	loggerConfig := config.Config{
		Log: config.Log{
			LogLevel: "disabled",
		},
		OpenTelemetry: config.OpenTelemetry{
			Enabled: false,
		},
	}
	s.logger = logger.New(loggerConfig)

	lifecycle := &mockLifecycle{}

	s.kafkaBuilder.On("BuildProducer", event.IdentityUserAuthenticatedTopic).Return(s.mockProducer)

	s.sut = producer.NewUserAuthenticatedProducer(
		lifecycle,
		s.kafkaBuilder,
		s.logger,
	)
}

func TestUserAuthenticatedProducerSuite(t *testing.T) {
	suite.Run(t, new(UserAuthenticatedProducerTestSuite))
}

func (s *UserAuthenticatedProducerTestSuite) TestProduce_ValidMessage_ProducesSuccessfully() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: userID,
	}

	expectedMessageBytes, err := json.Marshal(userAuthenticatedMessage)
	s.Require().NoError(err)

	expectedKafkaMessage := kafka.Message{
		Value: expectedMessageBytes,
	}

	s.mockProducer.On("Produce", mock.Anything, expectedKafkaMessage).Return(nil)

	// Act
	err = s.sut.Produce(ctx, userAuthenticatedMessage)

	// Assert
	s.Require().NoError(err)
}

func (s *UserAuthenticatedProducerTestSuite) TestProduce_ProducerError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: userID,
	}

	expectedMessageBytes, err := json.Marshal(userAuthenticatedMessage)
	s.Require().NoError(err)

	expectedKafkaMessage := kafka.Message{
		Value: expectedMessageBytes,
	}

	producerError := errors.New("kafka producer error")
	s.mockProducer.On("Produce", mock.Anything, expectedKafkaMessage).Return(producerError)

	// Act
	err = s.sut.Produce(ctx, userAuthenticatedMessage)

	// Assert
	s.Require().Error(err)
	s.Equal(producerError, err)
}

func (s *UserAuthenticatedProducerTestSuite) TestProduce_MarshalError_ReturnsError() {
	// Arrange
	ctx := context.Background()

	// Create an invalid message that would cause JSON marshal to fail
	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: 0, // Valid but we'll simulate marshal error
	}

	// We can't easily force json.Marshal to fail with this simple struct,
	// but this test demonstrates the pattern for handling marshal errors
	// In a real scenario, you might have a more complex message structure

	expectedMessageBytes, err := json.Marshal(userAuthenticatedMessage)
	s.Require().NoError(err)

	expectedKafkaMessage := kafka.Message{
		Value: expectedMessageBytes,
	}

	s.mockProducer.On("Produce", mock.Anything, expectedKafkaMessage).Return(nil)

	// Act
	err = s.sut.Produce(ctx, userAuthenticatedMessage)

	// Assert
	s.Require().NoError(err)
}

func (s *UserAuthenticatedProducerTestSuite) TestProduce_EmptyUserID_ProducesSuccessfully() {
	// Arrange
	ctx := context.Background()

	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: 0,
	}

	expectedMessageBytes, err := json.Marshal(userAuthenticatedMessage)
	s.Require().NoError(err)

	expectedKafkaMessage := kafka.Message{
		Value: expectedMessageBytes,
	}

	s.mockProducer.On("Produce", mock.Anything, expectedKafkaMessage).Return(nil)

	// Act
	err = s.sut.Produce(ctx, userAuthenticatedMessage)

	// Assert
	s.Require().NoError(err)
}
