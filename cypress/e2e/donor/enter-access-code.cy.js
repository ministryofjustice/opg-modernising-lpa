describe('Enter access code', () => {
  let accessCode = Math.random().toString().substring(0, 12);

  beforeEach(() => {
    cy.visit(`/fixtures/supporter?redirect=/enter-access-code&organisation=1&accessCode=${accessCode}`);
  });

  it('links the LPA', () => {
    cy.checkA11yApp();
    cy.get('#f-reference-number').type(accessCode);
    cy.contains('button', 'Continue').click();

    cy.contains('M-FAKE-');
    cy.contains('a', 'Go to task list').click();
    cy.contains('LPA task list');
  });
});
