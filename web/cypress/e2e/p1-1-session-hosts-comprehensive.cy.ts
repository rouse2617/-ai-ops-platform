import { ChatPage } from '../support/page-objects'

describe('P1-1: Session Host Association - Comprehensive', () => {
  let chatPage: ChatPage

  beforeEach(() => {
    chatPage = new ChatPage()
    chatPage.visit()
  })

  describe('TC01: New session associates with selected hosts', () => {
    it('should associate selected hosts with new session', () => {
      chatPage.selectHosts(['host-001', 'host-002'])
      chatPage.getNewSessionButton().click()

      chatPage.getSessionHosts().should('contain', 'host-001')
      chatPage.getSessionHosts().should('contain', 'host-002')
    })

    it('should display host count in session', () => {
      chatPage.selectHosts(['host-001', 'host-002', 'host-003'])
      chatPage.getNewSessionButton().click()

      cy.get('[data-testid="session-host-count"]').should('contain', '3')
    })

    it('should persist hosts in session data', () => {
      chatPage.selectHosts(['host-001', 'host-002'])
      chatPage.getNewSessionButton().click()

      cy.get('[data-testid="session-data"]').should('contain', 'host-001')
      cy.get('[data-testid="session-data"]').should('contain', 'host-002')
    })
  })

  describe('TC02: Host changes persist after refresh', () => {
    it('should display updated hosts after page refresh', () => {
      chatPage.selectHosts(['host-001', 'host-002'])
      chatPage.getNewSessionButton().click()

      chatPage.getSessionHosts().should('contain', 'host-001')
      chatPage.getSessionHosts().should('contain', 'host-002')

      cy.reload()

      chatPage.getSessionHosts().should('contain', 'host-001')
      chatPage.getSessionHosts().should('contain', 'host-002')
    })

    it('should restore session from localStorage', () => {
      chatPage.selectHosts(['host-001', 'host-002'])
      chatPage.getNewSessionButton().click()

      const sessionId = cy.get('[data-testid="current-session-id"]').invoke('text')

      cy.reload()

      cy.get('[data-testid="current-session-id"]').invoke('text').should('equal', sessionId)
    })

    it('should maintain host order after refresh', () => {
      chatPage.selectHosts(['host-003', 'host-001', 'host-002'])
      chatPage.getNewSessionButton().click()

      cy.reload()

      cy.get('[data-testid="session-host"]').eq(0).should('contain', 'host-003')
      cy.get('[data-testid="session-host"]').eq(1).should('contain', 'host-001')
      cy.get('[data-testid="session-host"]').eq(2).should('contain', 'host-002')
    })
  })

  describe('TC03: Switching sessions shows correct hosts', () => {
    it('should display corresponding hosts when switching sessions', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.getNewSessionButton().click()
      chatPage.getSessionHosts().should('contain', 'host-001')

      chatPage.selectHosts(['host-002', 'host-003'])
      chatPage.getNewSessionButton().click()
      chatPage.getSessionHosts().should('contain', 'host-002')
      chatPage.getSessionHosts().should('contain', 'host-003')

      chatPage.getSessionItem(0).click()
      chatPage.getSessionHosts().should('contain', 'host-001')
      chatPage.getSessionHosts().should('not.contain', 'host-002')
    })

    it('should update host selector when switching sessions', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.getNewSessionButton().click()

      chatPage.selectHosts(['host-002', 'host-003'])
      chatPage.getNewSessionButton().click()

      chatPage.getSessionItem(0).click()
      cy.get('[data-testid="host-selector"]').should('contain', 'host-001')
    })

    it('should preserve session history when switching', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.getNewSessionButton().click()
      chatPage.sendMessage('查看系统负载')

      chatPage.selectHosts(['host-002'])
      chatPage.getNewSessionButton().click()
      chatPage.sendMessage('查看进程列表')

      chatPage.getSessionItem(0).click()
      chatPage.getMessageList().should('contain', '查看系统负载')
      chatPage.getMessageList().should('not.contain', '查看进程列表')
    })
  })

  describe('Session host management', () => {
    it('should allow updating hosts in existing session', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.getNewSessionButton().click()

      chatPage.selectHosts(['host-002', 'host-003'])
      cy.get('[data-testid="update-session-hosts"]').click()

      chatPage.getSessionHosts().should('contain', 'host-002')
      chatPage.getSessionHosts().should('contain', 'host-003')
    })

    it('should display host status in session', () => {
      chatPage.selectHosts(['host-001', 'host-002'])
      chatPage.getNewSessionButton().click()

      cy.get('[data-testid="session-host-status"]').eq(0).should('have.class', 'online')
      cy.get('[data-testid="session-host-status"]').eq(1).should('have.class', 'offline')
    })
  })

  describe('Multiple sessions with different hosts', () => {
    it('should manage multiple sessions independently', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.getNewSessionButton().click()
      const session1Id = cy.get('[data-testid="current-session-id"]').invoke('text')

      chatPage.selectHosts(['host-002', 'host-003'])
      chatPage.getNewSessionButton().click()
      const session2Id = cy.get('[data-testid="current-session-id"]').invoke('text')

      session1Id.then(id1 => {
        session2Id.then(id2 => {
          expect(id1).not.to.equal(id2)
        })
      })
    })
  })
})
