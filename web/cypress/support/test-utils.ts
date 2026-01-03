// Test utilities and helpers

export function generateRandomHostId(): string {
  return `host-${String(Math.floor(Math.random() * 1000)).padStart(3, '0')}`
}

export function generateRandomSessionId(): string {
  return `session-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
}

export function waitForElement(selector: string, timeout = 5000): Cypress.Chainable<JQuery<HTMLElement>> {
  return cy.get(selector, { timeout })
}

export function waitForText(text: string, timeout = 5000): Cypress.Chainable<JQuery<HTMLElement>> {
  return cy.contains(text, { timeout })
}

export function scrollToElement(selector: string): void {
  cy.get(selector).scrollIntoView()
}

export function takeScreenshot(name: string): void {
  cy.screenshot(name, { overwrite: true })
}

export function clearLocalStorage(): void {
  cy.window().then(win => {
    win.localStorage.clear()
  })
}

export function getLocalStorageItem(key: string): Cypress.Chainable<string | null> {
  return cy.window().then(win => {
    return win.localStorage.getItem(key)
  })
}

export function setLocalStorageItem(key: string, value: string): void {
  cy.window().then(win => {
    win.localStorage.setItem(key, value)
  })
}
