describe('P0-1: High-Risk Operation Confirmation', () => {
  beforeEach(() => {
    cy.visit('/')
  })

  describe('TC01: rm -rf command triggers confirmation dialog', () => {
    it('should show confirmation dialog for rm -rf /tmp/test', () => {
      cy.sendMessage('帮我执行 rm -rf /tmp/test')
      cy.getConfirmDialog().should('be.visible')
      cy.get('[data-testid="dialog-title"]').should('contain', '危险操作确认')
      cy.get('[data-testid="risk-level"]').should('contain', '高风险操作')
    })
  })

  describe('TC02: mkfs command requires CONFIRM text input', () => {
    it('should require CONFIRM text for mkfs.ext4 /dev/sdb1', () => {
      cy.sendMessage('执行 mkfs.ext4 /dev/sdb1')
      cy.getConfirmDialog().should('be.visible')
      cy.get('[data-testid="risk-level"]').should('contain', '极高风险操作')
      cy.get('[data-testid="confirm-input"]').should('be.visible')
      cy.get('[data-testid="confirm-button"]').should('be.disabled')

      cy.get('[data-testid="danger-checkbox"]').click()
      cy.get('[data-testid="confirm-button"]').should('be.disabled')

      cy.get('[data-testid="confirm-input"]').type('CONFIRM')
      cy.get('[data-testid="confirm-button"]').should('not.be.disabled')
    })
  })

  describe('TC03: Cancel button closes dialog without sending message', () => {
    it('should close dialog and not send message when clicking cancel', () => {
      cy.sendMessage('帮我执行 rm -rf /tmp/test')
      cy.getConfirmDialog().should('be.visible')

      cy.cancelDangerOperation()
      cy.getConfirmDialog().should('not.exist')

      cy.get('[data-testid="message-list"]').within(() => {
        cy.get('[data-testid="message"]').last().should('contain', '帮我执行 rm -rf /tmp/test')
      })
    })
  })

  describe('TC04: Confirm button after countdown sends message', () => {
    it('should send message after confirming dangerous operation', () => {
      cy.sendMessage('帮我执行 rm -rf /tmp/test')
      cy.getConfirmDialog().should('be.visible')

      cy.get('[data-testid="danger-checkbox"]').click()
      cy.confirmDangerOperation()

      cy.getConfirmDialog().should('not.exist')
      cy.get('[data-testid="message-list"]').should('contain', '执行中')
    })
  })

  describe('TC05: Safe commands do not trigger confirmation dialog', () => {
    it('should not show dialog for safe commands like "查看系统负载"', () => {
      cy.sendMessage('查看系统负载')
      cy.getConfirmDialog().should('not.exist')
      cy.get('[data-testid="message-list"]').should('contain', '查看系统负载')
    })
  })

  describe('TC06: ESC key closes confirmation dialog', () => {
    it('should close dialog when pressing ESC', () => {
      cy.sendMessage('帮我执行 rm -rf /tmp/test')
      cy.getConfirmDialog().should('be.visible')

      cy.get('body').type('{esc}')
      cy.getConfirmDialog().should('not.exist')
    })
  })
})
