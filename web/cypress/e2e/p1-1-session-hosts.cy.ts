describe('P1-1: Session Host Association', () => {
  beforeEach(() => {
    cy.visit('/')
  })

  describe('TC01: New session associates with selected hosts', () => {
    it('should associate selected hosts with new session', () => {
      cy.selectHosts(['host-001', 'host-002'])
      cy.get('[data-testid="new-session-button"]').click()

      cy.get('[data-testid="session-hosts"]').should('contain', 'host-001')
      cy.get('[data-testid="session-hosts"]').should('contain', 'host-002')
    })
  })

  describe('TC02: Host changes persist after refresh', () => {
    it('should display updated hosts after page refresh', () => {
      cy.selectHosts(['host-001', 'host-002'])
      cy.get('[data-testid="new-session-button"]').click()

      cy.get('[data-testid="session-hosts"]').should('contain', 'host-001')

      cy.reload()
      cy.get('[data-testid="session-hosts"]').should('contain', 'host-001')
      cy.get('[data-testid="session-hosts"]').should('contain', 'host-002')
    })
  })

  describe('TC03: Switching sessions shows correct hosts', () => {
    it('should display corresponding hosts when switching sessions', () => {
      cy.selectHosts(['host-001'])
      cy.get('[data-testid="new-session-button"]').click()
      cy.get('[data-testid="session-hosts"]').should('contain', 'host-001')

      cy.selectHosts(['host-002', 'host-003'])
      cy.get('[data-testid="new-session-button"]').click()
      cy.get('[data-testid="session-hosts"]').should('contain', 'host-002')
      cy.get('[data-testid="session-hosts"]').should('contain', 'host-003')

      cy.get('[data-testid="session-list"]').within(() => {
        cy.get('[data-testid="session-item"]').first().click()
      })
      cy.get('[data-testid="session-hosts"]').should('contain', 'host-001')
    })
  })
})
