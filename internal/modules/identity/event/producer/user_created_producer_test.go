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
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	kafka_mocks "github.com/cristiano-pacheco/pingo/pkg/kafka/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UserCreatedProducerTestSuite struct {
	suite.Suite
	sut          producer.UserCreatedProducer
	producerMock *kafka_mocks.MockProducer
	builderMock  *kafka_mocks.MockBuilder
	logger       logger.Logger
	otel         otel.Otel
}

func (s *UserCreatedProducerTestSuite) SetupTest() {
	s.producerMock = kafka_mocks.NewMockProducer(s.T())
	s.builderMock = kafka_mocks.NewMockBuilder(s.T())
	s.otel = otel.NewNoopOtel()

	loggerConfig := config.Config{
		Log: config.Log{
			LogLevel: "disabled",
		},
		OpenTelemetry: config.OpenTelemetry{
			Enabled: false,
		},
	}
	s.logger = logger.New(loggerConfig)

	s.builderMock.On("BuildProducer", event.IdentityUserCreatedTopic).Return(s.producerMock)

	lifecycle := &mockLifecycle{}
	s.sut = producer.NewUserCreatedProducer(
		lifecycle,
		s.logger,
		s.otel,
		s.builderMock,
	)
}

func TestUserCreatedProducerSuite(t *testing.T) {
	suite.Run(t, new(UserCreatedProducerTestSuite))
}

func (s *UserCreatedProducerTestSuite) TestProduce_ValidMessage_ProducesSuccessfully() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userCreatedMessage := event.UserCreatedMessage{
		UserID: userID,
	}

	expectedMessageBytes, err := json.Marshal(userCreatedMessage)
	s.Require().NoError(err)

	expectedKafkaMessage := kafka.Message{
		Value: expectedMessageBytes,
	}

	s.producerMock.On("Produce", mock.Anything, expectedKafkaMessage).Return(nil)

	// Act
	err = s.sut.Produce(ctx, userCreatedMessage)

	// Assert
	s.Require().NoError(err)
}

func (s *UserCreatedProducerTestSuite) TestProduce_ProducerFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userCreatedMessage := event.UserCreatedMessage{
		UserID: userID,
	}

	expectedMessageBytes, err := json.Marshal(userCreatedMessage)
	s.Require().NoError(err)

	expectedKafkaMessage := kafka.Message{
		Value: expectedMessageBytes,
	}

	producerError := errors.New("failed to produce message")
	s.producerMock.On("Produce", mock.Anything, expectedKafkaMessage).Return(producerError)

	// Act
	err = s.sut.Produce(ctx, userCreatedMessage)

	// Assert
	s.Require().Error(err)
	s.Equal(producerError, err)
}

func (s *UserCreatedProducerTestSuite) TestProduce_ZeroUserID_ProducesSuccessfully() {
	// Arrange
	ctx := context.Background()
	userID := uint64(0)

	userCreatedMessage := event.UserCreatedMessage{
		UserID: userID,
	}

	expectedMessageBytes, err := json.Marshal(userCreatedMessage)
	s.Require().NoError(err)

	expectedKafkaMessage := kafka.Message{
		Value: expectedMessageBytes,
	}

	s.producerMock.On("Produce", mock.Anything, expectedKafkaMessage).Return(nil)

	// Act
	err = s.sut.Produce(ctx, userCreatedMessage)

	// Assert
	s.Require().NoError(err)
}

func (s *UserCreatedProducerTestSuite) TestProduce_LargeUserID_ProducesSuccessfully() {
	// Arrange
	ctx := context.Background()
	userID := uint64(9223372036854775807) // Max int64 value

	userCreatedMessage := event.UserCreatedMessage{
		UserID: userID,
	}

	expectedMessageBytes, err := json.Marshal(userCreatedMessage)
	s.Require().NoError(err)

	expectedKafkaMessage := kafka.Message{
		Value: expectedMessageBytes,
	}

	s.producerMock.On("Produce", mock.Anything, expectedKafkaMessage).Return(nil)

	// Act
	err = s.sut.Produce(ctx, userCreatedMessage)

	// Assert
	s.Require().NoError(err)
}
