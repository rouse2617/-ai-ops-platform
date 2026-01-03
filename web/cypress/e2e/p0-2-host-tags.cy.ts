describe('P0-2: Host Tags Display', () => {
  beforeEach(() => {
    cy.visit('/')
  })

  describe('TC01: Single host execution shows host tag', () => {
    it('should display host tag when executing on 1 host', () => {
      cy.selectHosts(['host-001'])
      cy.sendMessage('查看系统负载')

      cy.get('[data-testid="message-list"]').within(() => {
        cy.get('[data-testid="host-tag"]').should('have.length', 1)
        cy.get('[data-testid="host-tag"]').should('contain', 'host-001')
      })
    })
  })

  describe('TC02: Multiple hosts execution shows multiple tags', () => {
    it('should display 3 host tags when executing on 3 hosts', () => {
      cy.selectHosts(['host-001', 'host-002', 'host-003'])
      cy.sendMessage('查看系统负载')

      cy.get('[data-testid="message-list"]').within(() => {
        cy.get('[data-testid="host-tag"]').should('have.length', 3)
        cy.get('[data-testid="host-tag"]').eq(0).should('contain', 'host-001')
        cy.get('[data-testid="host-tag"]').eq(1).should('contain', 'host-002')
        cy.get('[data-testid="host-tag"]').eq(2).should('contain', 'host-003')
      })
    })
  })

  describe('TC03: Offline host shows red status', () => {
    it('should display failed host in red when one host is offline', () => {
      cy.selectHosts(['host-001', 'host-002', 'host-003'])
      cy.sendMessage('查看系统负载')

      cy.get('[data-testid="message-list"]').within(() => {
        cy.get('[data-testid="host-tag"]').eq(1).should('have.class', 'status-failed')
        cy.get('[data-testid="host-tag"]').eq(1).should('have.css', 'background-color', 'rgb(245, 108, 108)')
      })
    })
  })

  describe('TC04: Collapse button toggles result area', () => {
    it('should collapse and expand result area when clicking collapse button', () => {
      cy.selectHosts(['host-001'])
      cy.sendMessage('查看系统负载')

      cy.get('[data-testid="result-area"]').should('be.visible')
      cy.get('[data-testid="collapse-button"]').click()
      cy.get('[data-testid="result-area"]').should('not.be.visible')

      cy.get('[data-testid="collapse-button"]').click()
      cy.get('[data-testid="result-area"]').should('be.visible')
    })
  })
})
