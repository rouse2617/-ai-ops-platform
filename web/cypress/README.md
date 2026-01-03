# E2E Test Suite Documentation

## Overview

This directory contains comprehensive end-to-end tests for the AI Pro application using Cypress. The tests cover all P0 and P1 priority features.

## Test Structure

```
cypress/
├── e2e/                          # Test files
│   ├── p0-1-danger-*.cy.ts      # High-risk operation confirmation tests
│   ├── p0-2-host-tags-*.cy.ts   # Host tags display tests
│   ├── p1-1-session-hosts-*.cy.ts # Session host association tests
│   ├── p1-2-panel-sync-*.cy.ts  # Panel synchronization tests
│   └── p1-3-long-content-*.cy.ts # Long content collapsing tests
├── support/
│   ├── commands.ts              # Custom Cypress commands
│   ├── page-objects.ts          # Page Object Models
│   ├── test-config.ts           # Test configuration
│   ├── test-utils.ts            # Utility functions
│   └── e2e.ts                   # E2E setup
├── fixtures/
│   ├── messages.json            # Sample messages
│   └── hosts.json               # Sample hosts
└── cypress.config.ts            # Cypress configuration
```

## Running Tests

### Run all E2E tests
```bash
npm run test:e2e
```

### Run tests in interactive mode
```bash
npm run test:e2e:open
```

### Run P0 priority tests only
```bash
npm run test:e2e:p0
```

### Run P1 priority tests only
```bash
npm run test:e2e:p1
```

### Run specific test file
```bash
npx cypress run --spec "cypress/e2e/p0-1-danger-comprehensive.cy.ts"
```

### Run tests with specific browser
```bash
npx cypress run --browser chrome
npx cypress run --browser firefox
npx cypress run --browser edge
```

## Test Coverage

### P0-1: High-Risk Operation Confirmation
- **TC01**: rm -rf command triggers confirmation dialog
- **TC02**: mkfs command requires CONFIRM text input
- **TC03**: Cancel button closes dialog without sending message
- **TC04**: Confirm button sends message after confirmation
- **TC05**: Safe commands do not trigger confirmation dialog
- **TC06**: ESC key closes confirmation dialog

### P0-2: Host Tags Display
- **TC01**: Single host execution shows host tag
- **TC02**: Multiple hosts execution shows multiple tags
- **TC03**: Offline host shows red status
- **TC04**: Collapse button toggles result area

### P1-1: Session Host Association
- **TC01**: New session associates with selected hosts
- **TC02**: Host changes persist after refresh
- **TC03**: Switching sessions shows correct hosts

### P1-2: Panel Synchronization
- **TC01**: System load message syncs to right panel
- **TC02**: Log analysis message syncs to right panel
- **TC03**: Process monitoring message syncs to right panel
- **TC04**: Network status message syncs to right panel

### P1-3: Long Content Collapsing
- **TC01**: Output over 50 lines shows collapse
- **TC02**: Expand button shows full content
- **TC03**: Collapse button restores collapsed state
- **TC04**: Short content does not show collapse button

## Custom Commands

### selectHosts(hostIds: string[])
Select multiple hosts for execution.
```typescript
cy.selectHosts(['host-001', 'host-002'])
```

### sendMessage(message: string)
Send a message in the chat.
```typescript
cy.sendMessage('查看系统负载')
```

### getConfirmDialog()
Get the danger confirmation dialog element.
```typescript
cy.getConfirmDialog().should('be.visible')
```

### confirmDangerOperation(confirmText?: string)
Confirm a dangerous operation.
```typescript
cy.confirmDangerOperation('CONFIRM')
```

### cancelDangerOperation()
Cancel a dangerous operation.
```typescript
cy.cancelDangerOperation()
```

## Page Object Models

### ChatPage
Main chat interface page object.
```typescript
const chatPage = new ChatPage()
chatPage.visit()
chatPage.selectHosts(['host-001'])
chatPage.sendMessage('查看系统负载')
```

### DangerConfirmPage
Danger confirmation dialog page object.
```typescript
const dangerPage = new DangerConfirmPage()
dangerPage.isVisible()
dangerPage.confirm('CONFIRM')
```

### CollapsibleOutputPage
Collapsible output page object.
```typescript
const outputPage = new CollapsibleOutputPage()
outputPage.toggleCollapse()
outputPage.isExpanded()
```

## Test Data

### Test Hosts
- `host-001`: Online server
- `host-002`: Offline server
- `host-003`: Online server

### Dangerous Commands
- `rm -rf /tmp/test`
- `mkfs.ext4 /dev/sdb1`
- `reboot`
- `shutdown -h now`
- `kill -9 1`

### Safe Commands
- `查看系统负载`
- `查看进程列表`
- `查看网络状态`
- `分析系统日志`

## Best Practices

### 1. Use Page Objects
Always use page objects for better maintainability:
```typescript
const chatPage = new ChatPage()
chatPage.sendMessage('test')
```

### 2. Wait for Elements
Use proper wait strategies:
```typescript
cy.get('[data-testid="element"]', { timeout: 10000 }).should('exist')
```

### 3. Use Data Attributes
Always use `data-testid` attributes for selectors:
```typescript
cy.get('[data-testid="send-button"]').click()
```

### 4. Avoid Hard Waits
Use Cypress built-in waiting instead of `cy.wait()`:
```typescript
cy.get('[data-testid="message"]').should('contain', 'text')
```

### 5. Test Isolation
Each test should be independent:
```typescript
beforeEach(() => {
  chatPage.visit()
  // Reset state
})
```

## Debugging Tests

### View test execution
```bash
npm run test:e2e:open
```

### Run single test
```bash
npx cypress run --spec "cypress/e2e/p0-1-danger-comprehensive.cy.ts" --headed
```

### Debug mode
```bash
npx cypress run --spec "cypress/e2e/p0-1-danger-comprehensive.cy.ts" --headed --no-exit
```

### View console logs
```typescript
cy.window().then(win => {
  console.log(win.console.log)
})
```

## CI/CD Integration

### GitHub Actions
Add to `.github/workflows/e2e.yml`:
```yaml
- name: Run E2E tests
  run: npm run test:e2e
```

### Environment Variables
Set in CI/CD:
```
CYPRESS_BASE_URL=http://localhost:1281
CYPRESS_API_URL=http://localhost:1280/api
```

## Troubleshooting

### Tests timing out
- Increase timeout in `cypress.config.ts`
- Check if backend is running
- Verify network connectivity

### Element not found
- Check `data-testid` attributes in components
- Verify element is visible before interaction
- Use `cy.debug()` to inspect state

### Flaky tests
- Add proper wait conditions
- Avoid hard waits
- Use `cy.intercept()` for API calls
- Ensure test isolation

## Contributing

When adding new tests:
1. Follow the existing test structure
2. Use page objects for UI interactions
3. Add descriptive test names
4. Include both positive and negative cases
5. Update this documentation

## Resources

- [Cypress Documentation](https://docs.cypress.io)
- [Best Practices](https://docs.cypress.io/guides/references/best-practices)
- [API Reference](https://docs.cypress.io/api/table-of-contents)
