describe('Dashboard', () => {
  beforeEach(() => {
    cy.visit('/fixtures/supporter?redirect=/supporter-dashboard&organisation=1&lpa=1');
  });

  it('shows LPAs', () => {
    cy.contains('Sam Smith');
    cy.contains('B14 7ED');
    cy.contains('Property and affairs');
    cy.contains('In progress');

    cy.contains('a', 'M-').click();
    cy.contains('Provide your details').click();
    cy.get('#f-first-names').type('2');
    cy.contains('button', 'Continue').click();
    cy.contains('a', 'Dashboard').click();
    cy.contains('Sam2 Smith');
  });

  it('can create a new LPA', () => {
    cy.checkA11yApp();
    cy.contains('button', 'Make a new LPA').click();

    cy.get('#f-first-names').type('John');
    cy.get('#f-last-name').type('Doe');
    cy.get('#f-date-of-birth').type('1');
    cy.get('#f-date-of-birth-month').type('2');
    cy.get('#f-date-of-birth-year').type('1990');
    cy.get('#f-can-sign').check();
    cy.contains('button', 'Continue').click();
    cy.contains('a', 'Dashboard').click();

    cy.contains('John Doe');
  });
});
