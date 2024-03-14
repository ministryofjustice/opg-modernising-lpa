describe('Dashboard', () => {
  beforeEach(() => {
    cy.visit('/fixtures/supporter?redirect=/dashboard&organisation=1&lpa=1');
  });

  it('shows LPAs', () => {
    cy.checkA11yApp();

    cy.contains('Sam Smith');
    cy.contains('B14 7ED');
    cy.contains('Property and affairs');
    cy.contains('In progress');

    cy.contains('a', 'M-');
  });

  it('can start a new LPA', () => {
    cy.contains('Cymraeg').should('not.exist');
    cy.contains('a', 'Make a new LPA').click();

    cy.checkA11yApp();
    cy.contains('Cymraeg').should('not.exist');
    cy.contains('label', 'Make an online LPA').click();
    cy.contains('button', 'Continue').click();

    cy.contains('Cymraeg');
    cy.get('#f-first-names').type('John');
    cy.get('#f-last-name').type('Doe');
    cy.get('#f-date-of-birth').type('1');
    cy.get('#f-date-of-birth-month').type('2');
    cy.get('#f-date-of-birth-year').type('1990');
    cy.get('#f-can-sign').check({ force: true });
    cy.contains('button', 'Continue').click();
  });

  it('can show guidance for starting a paper LPA', () => {
    cy.contains('a', 'Make a new LPA').click();
    cy.contains('label', 'Offline').click();
    cy.contains('button', 'Continue').click();

    cy.checkA11yApp();
  });
});
