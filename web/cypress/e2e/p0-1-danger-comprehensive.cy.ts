import { ChatPage, DangerConfirmPage, CollapsibleOutputPage } from './page-objects'

describe('P0-1: High-Risk Operation Confirmation - Comprehensive', () => {
  let chatPage: ChatPage
  let dangerPage: DangerConfirmPage

  beforeEach(() => {
    chatPage = new ChatPage()
    dangerPage = new DangerConfirmPage()
    chatPage.visit()
  })

  describe('TC01: rm -rf command triggers confirmation dialog', () => {
    it('should show confirmation dialog with correct risk level', () => {
      chatPage.sendMessage('帮我执行 rm -rf /tmp/test')

      dangerPage.isVisible()
      dangerPage.getDialogTitle().should('contain', '危险操作确认')
      dangerPage.getRiskLevel().should('contain', '高风险操作')
      dangerPage.getAffectedHosts().should('exist')
    })

    it('should display command in dialog', () => {
      chatPage.sendMessage('帮我执行 rm -rf /tmp/test')

      dangerPage.getDialog().should('contain', 'rm -rf /tmp/test')
    })

    it('should disable confirm button until checkbox is checked', () => {
      chatPage.sendMessage('帮我执行 rm -rf /tmp/test')

      dangerPage.getConfirmButton().should('be.disabled')
      dangerPage.getConfirmCheckbox().click()
      dangerPage.getConfirmButton().should('not.be.disabled')
    })
  })

  describe('TC02: mkfs command requires CONFIRM text input', () => {
    it('should show critical risk level for mkfs command', () => {
      chatPage.sendMessage('执行 mkfs.ext4 /dev/sdb1')

      dangerPage.isVisible()
      dangerPage.getRiskLevel().should('contain', '极高风险操作')
    })

    it('should require CONFIRM text for critical operations', () => {
      chatPage.sendMessage('执行 mkfs.ext4 /dev/sdb1')

      dangerPage.getConfirmInput().should('be.visible')
      dangerPage.getConfirmButton().should('be.disabled')

      dangerPage.getConfirmCheckbox().click()
      dangerPage.getConfirmButton().should('be.disabled')

      dangerPage.getConfirmInput().type('CONFIRM')
      dangerPage.getConfirmButton().should('not.be.disabled')
    })

    it('should reject incorrect confirmation text', () => {
      chatPage.sendMessage('执行 mkfs.ext4 /dev/sdb1')

      dangerPage.getConfirmCheckbox().click()
      dangerPage.getConfirmInput().type('confirm')
      dangerPage.getConfirmButton().should('be.disabled')

      dangerPage.getConfirmInput().clear()
      dangerPage.getConfirmInput().type('CONFIRM')
      dangerPage.getConfirmButton().should('not.be.disabled')
    })
  })

  describe('TC03: Cancel button closes dialog without sending message', () => {
    it('should close dialog when clicking cancel', () => {
      chatPage.sendMessage('帮我执行 rm -rf /tmp/test')
      dangerPage.isVisible()

      dangerPage.cancel()
      dangerPage.isHidden()
    })

    it('should not send message when canceling', () => {
      const messageCountBefore = cy.get('[data-testid="message"]').its('length')

      chatPage.sendMessage('帮我执行 rm -rf /tmp/test')
      dangerPage.cancel()

      cy.get('[data-testid="message"]').its('length').should('equal', messageCountBefore)
    })
  })

  describe('TC04: Confirm button sends message after confirmation', () => {
    it('should send message after confirming dangerous operation', () => {
      chatPage.sendMessage('帮我执行 rm -rf /tmp/test')

      dangerPage.confirm()
      dangerPage.isHidden()

      cy.get('[data-testid="message-list"]').should('contain', '执行中')
    })

    it('should show loading state after confirmation', () => {
      chatPage.sendMessage('帮我执行 rm -rf /tmp/test')

      dangerPage.confirm()
      cy.get('[data-testid="loading-indicator"]').should('be.visible')
    })
  })

  describe('TC05: Safe commands do not trigger confirmation dialog', () => {
    it('should not show dialog for safe commands', () => {
      chatPage.sendMessage('查看系统负载')

      dangerPage.isHidden()
      chatPage.getLastMessage().should('contain', '查看系统负载')
    })

    it('should not show dialog for read-only commands', () => {
      chatPage.sendMessage('查看进程列表')
      dangerPage.isHidden()

      chatPage.sendMessage('查看网络状态')
      dangerPage.isHidden()
    })
  })

  describe('TC06: ESC key closes confirmation dialog', () => {
    it('should close dialog when pressing ESC', () => {
      chatPage.sendMessage('帮我执行 rm -rf /tmp/test')
      dangerPage.isVisible()

      cy.get('body').type('{esc}')
      dangerPage.isHidden()
    })

    it('should not send message when closing with ESC', () => {
      chatPage.sendMessage('帮我执行 rm -rf /tmp/test')
      cy.get('body').type('{esc}')

      cy.get('[data-testid="message-list"]').should('not.contain', '执行中')
    })
  })

  describe('Multiple dangerous commands in sequence', () => {
    it('should handle multiple dangerous operations correctly', () => {
      chatPage.sendMessage('帮我执行 rm -rf /tmp/test')
      dangerPage.isVisible()
      dangerPage.cancel()

      chatPage.sendMessage('执行 mkfs.ext4 /dev/sdb1')
      dangerPage.isVisible()
      dangerPage.getRiskLevel().should('contain', '极高风险操作')
      dangerPage.cancel()

      chatPage.sendMessage('查看系统负载')
      dangerPage.isHidden()
    })
  })
})
