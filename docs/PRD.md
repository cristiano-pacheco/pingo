# Pingo - Product Requirements Document

## 1. Product Overview

**Product Name:** Pingo  
**Version:** 1.0  
**Type:** Self-hosted, open-source monitoring tool  
**Target Audience:** Developers, DevOps engineers, small to medium businesses  

### 1.1 Product Vision
Pingo is a self-hosted, open-source monitoring tool that provides basic uptime monitoring with email and webhook notifications. It serves as a learning-focused alternative to commercial solutions like Pingdom, showcasing modern development practices while delivering fully functional monitoring capabilities.

### 1.2 Key Value Propositions
- **Self-hosted**: Complete control over data and infrastructure
- **Open-source**: Transparent, customizable, and community-driven
- **Learning-focused**: Clean codebase demonstrating modern development practices
- **Simple yet functional**: Essential monitoring features without complexity
- **Cost-effective**: Free alternative to commercial monitoring solutions

## 2. Core Features

### 2.1 Authentication System
- **Magic Link Login**: Passwordless authentication via email
- **User Registration**: Account creation with email confirmation
- **Account Management**: Profile updates and account settings

### 2.2 Monitoring Management
- **HTTP Monitor Creation**: Configure uptime checks for web services
- **Monitor Configuration**: Customizable timeouts, failure thresholds, and validation rules
- **Monitor Status Tracking**: Real-time status updates and historical data

### 2.3 Contact Management
- **Multiple Contact Types**: Support for email and webhook notifications
- **Contact Configuration**: Enable/disable contacts per monitor
- **Notification Routing**: Flexible assignment of contacts to monitors

### 2.4 Admin Interface
- **Dashboard**: Overview of all monitors and their status
- **Monitor Management**: CRUD operations for monitoring targets
- **Contact Management**: CRUD operations for notification contacts
- **User Profile**: Account settings and preferences

### 2.5 Alert Results Dashboard
- **Real-time Status Display**: Live monitoring status with visual indicators
- **Alert History**: Chronological view of all triggered alerts
- **Notification Log**: Track of all sent notifications with delivery status

## 3. User Stories

### 3.1 Authentication
- **As a new user**, I want to register for an account so that I can start monitoring my services
- **As a registered user**, I want to login using a magic link so that I don't need to remember passwords
- **As a user**, I want to confirm my email address so that I can receive notifications

### 3.2 Monitor Management
- **As a user**, I want to create HTTP monitors so that I can track my website uptime
- **As a user**, I want to configure monitor settings so that I can customize check frequency and failure criteria
- **As a user**, I want to enable/disable monitors so that I can control which services are being monitored

### 3.3 Contact Management
- **As a user**, I want to add email contacts so that I can receive notifications when services go down
- **As a user**, I want to add webhook contacts so that I can integrate with external systems
- **As a user**, I want to assign contacts to specific monitors so that I get relevant notifications

### 3.4 Monitoring Results Dashboard
- **As a user**, I want to view detailed monitoring results for each monitor so that I can analyze performance patterns
- **As a user**, I want to see historical uptime data so that I can track service reliability over time
- **As a user**, I want to view response time trends so that I can identify performance degradation

## 4. Form Fields Specification

### 4.1 User Management Forms

#### User Registration Form
- **name** (text, required): User's full name
- **email** (email, required): User's email address

#### User Profile Form
- **name** (text, required): User's full name
- **email** (email, required): User's email address
- **status** (select, required): Account status (active, inactive, pending)

### 4.2 Contact Management Forms

#### Contact Creation/Edit Form
- **name** (text, required): Display name for the contact
- **contact_type** (select, required): Type of contact (email, webhook)
- **contact_data** (text, required): Contact information (email address or webhook URL)
- **is_enabled** (checkbox): Whether the contact is active

### 4.3 Monitor Management Forms

#### HTTP Monitor Creation/Edit Form
- **name** (text, required): Display name for the monitor
- **http_url** (url, required): URL to monitor
- **http_method** (select, required): HTTP method (GET, POST, PUT, DELETE, HEAD)
- **check_timeout** (number, required): Timeout in seconds (1-300)
- **fail_threshold** (number, required): Number of failures before alerting (1-10)
- **valid_response_statuses** (text, required): Comma-separated list of valid HTTP status codes
- **request_headers** (textarea): JSON object of request headers
- **contact_ids** (multi-select): Associated contacts for notifications
- **is_enabled** (checkbox): Whether the monitor is active

## 5. Technical Requirements

### 5.1 Architecture Overview
The Pingo monitoring solution consists of two separate projects that work together:

#### 5.1.1 pingo-api (Backend)
- **Language**: Go (Golang)
- **Purpose**: REST API backend providing all business logic and data access
- **Database**: PostgreSQL with direct database access
- **Observability**: OpenTelemetry integration for monitoring and tracing
- **Responsibilities**:
  - User authentication and session management
  - Monitor configuration and execution
  - Contact management
  - Alert processing and notification delivery
  - Data persistence and retrieval
  - API endpoint exposure

#### 5.1.2 pingo-web (Frontend)
- **Language**: Go (Golang) 
- **Template Engine**: Templ for server-side rendering
- **CSS Framework**: Bootstrap or similar modern CSS framework
- **Purpose**: Web interface that renders the user-facing application
- **Data Access**: Consumes REST API from pingo-api (no direct database access)
- **Responsibilities**:
  - User interface rendering
  - Form handling and validation
  - Dashboard visualization
  - API consumption and data presentation
  - Session management (frontend)

### 5.2 Database Schema
The application uses PostgreSQL with the following core tables:
- **users**: User account information and authentication
- **contact**: Notification contacts (email/webhook)
- **http_monitor**: HTTP monitoring configuration

### 5.3 Authentication Flow
1. User enters email address in pingo-web
2. pingo-web calls pingo-api to generate and send magic link
3. User clicks link to authenticate
4. pingo-api validates token and creates session
5. pingo-web retrieves session info and redirects to dashboard

### 5.4 Monitoring Flow
1. pingo-api periodically checks configured HTTP endpoints
2. Evaluates response against configured criteria
3. Tracks failure count against threshold
4. Sends notifications when threshold is exceeded
5. Updates monitor status and logs results
6. pingo-web displays real-time status via API calls

### 5.5 API Communication
- **Protocol**: RESTful HTTP API
- **Data Format**: JSON
- **Authentication**: Session-based with secure tokens
- **Error Handling**: Standardized HTTP status codes and error responses
- **Rate Limiting**: Configurable limits to prevent abuse

## 6. User Interface Requirements

### 6.1 Dashboard
- Overview of all monitors with status indicators
- Quick access to create new monitors
- Recent activity log
- System status summary

### 6.2 Monitor Management
- List view of all monitors with status, last check time, and actions
- Create/edit forms with validation
- Detailed monitor configuration options
- Historical data visualization

### 6.3 Individual Monitor Dashboard
- **Monitor Overview**: Current status, uptime percentage, and key metrics
- **Response Time Chart**: Real-time and historical response time trends
- **Uptime Timeline**: Visual timeline showing up/down periods over selected timeframe
- **Check Results Log**: Detailed log of recent monitoring checks with timestamps and results
- **Incident History**: List of past incidents with duration, cause, and resolution details
- **Performance Metrics**: Average response time, success rate, and availability statistics
- **Alert Activity**: Recent alerts triggered for this monitor with notification status
- **Time Range Selector**: Filter data by last 24 hours, 7 days, 30 days, or custom range
- **Data Export**: Download monitoring data in CSV/JSON format for reporting

### 6.4 Contact Management
- List view of all contacts with type and status
- Create/edit forms with contact type selection
- Test functionality for webhooks
- Usage tracking (which monitors use each contact)

### 6.5 Responsive Design
- Mobile-friendly interface
- Progressive web app capabilities
- Accessible design following WCAG guidelines

## 7. Success Metrics

### 7.1 Functional Metrics
- Monitor uptime accuracy (>99.9%)
- Notification delivery success rate (>99%)
- System response time (<2 seconds)
- Zero data loss guarantee

### 7.2 User Experience Metrics
- User registration completion rate
- Time to first monitor setup (<5 minutes)
- User retention after 30 days
- Support ticket volume (target: <1% of user base)

## 8. Future Enhancements

### 8.1 Phase 2 Features
- SMS notifications
- Slack/Discord integrations
- Custom alert rules
- API endpoint monitoring
- SSL certificate monitoring

### 8.2 Phase 3 Features
- Team collaboration
- Advanced analytics and reporting
- Status page generation
- Multi-region monitoring
- Performance monitoring metrics

## 9. Constraints and Assumptions

### 9.1 Technical Constraints
- Self-hosted deployment model
- PostgreSQL database requirement
- Single-tenant architecture
- No external dependencies for core functionality

### 9.2 Business Constraints
- Open-source license (MIT/Apache 2.0)
- Community-driven development
- No commercial support obligation
- Educational/learning focus over enterprise features

### 9.3 Assumptions
- Users have basic technical knowledge for self-hosting
- Email delivery infrastructure is available
- Users primarily monitor HTTP/HTTPS endpoints
- Small to medium scale deployments (< 1000 monitors)