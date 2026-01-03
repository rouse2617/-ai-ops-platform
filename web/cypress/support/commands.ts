// Custom commands for E2E tests

Cypress.Commands.add('login', () => {
  cy.visit('/')
})

Cypress.Commands.add('selectHosts', (hostIds: string[]) => {
  cy.get('[data-testid="host-selector"]').click()
  hostIds.forEach(hostId => {
    cy.get(`[data-testid="host-option-${hostId}"]`).click()
  })
  cy.get('body').click(0, 0) // Close dropdown
})

Cypress.Commands.add('sendMessage', (message: string) => {
  cy.get('[data-testid="input-box"]').type(message)
  cy.get('[data-testid="send-button"]').click()
})

Cypress.Commands.add('waitForMessage', (timeout = 5000) => {
  cy.get('[data-testid="message-list"]', { timeout }).should('exist')
})

Cypress.Commands.add('getConfirmDialog', () => {
  return cy.get('[data-testid="danger-confirm-dialog"]')
})

Cypress.Commands.add('confirmDangerOperation', (confirmText?: string) => {
  cy.get('[data-testid="danger-checkbox"]').click()
  if (confirmText) {
    cy.get('[data-testid="confirm-input"]').type(confirmText)
  }
  cy.get('[data-testid="confirm-button"]').click()
})

Cypress.Commands.add('cancelDangerOperation', () => {
  cy.get('[data-testid="cancel-button"]').click()
})

declare global {
  namespace Cypress {
    interface Chainable {
      login(): Chainable<void>
      selectHosts(hostIds: string[]): Chainable<void>
      sendMessage(message: string): Chainable<void>
      waitForMessage(timeout?: number): Chainable<void>
      getConfirmDialog(): Chainable<JQuery<HTMLElement>>
      confirmDangerOperation(confirmText?: string): Chainable<void>
      cancelDangerOperation(): Chainable<void>
    }
  }
}

export {}
