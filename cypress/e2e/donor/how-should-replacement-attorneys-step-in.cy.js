describe('How should replacement attorneys step in', () => {
  beforeEach(() => {
    cy.visit('/fixtures?redirect=/how-should-replacement-attorneys-step-in&progress=chooseYourReplacementAttorneys&replacementAttorneys=single');
  });

  it('can choose how replacement attorneys step in', () => {
    cy.contains('h1', 'When your replacement attorneys step in');

    // see https://github.com/alphagov/govuk-frontend/issues/979
    cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

    cy.get('input[name="when-to-step-in"]').check('one-can-no-longer-act', { force: true });

    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/task-list');

    cy.checkA11yApp();
  });

  it('can choose how replacement attorneys step in - some other way', () => {
    cy.get('input[name="when-to-step-in"]').check('another-way', { force: true });
    cy.get('#f-other-details').type('some details on when to step in');

    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/task-list');
  });

  it('errors when unselected', () => {
    cy.contains('button', 'Save and continue').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Select when your replacement attorneys step in');
    });

    cy.contains('.govuk-fieldset .govuk-error-message', 'Select when your replacement attorneys step in');
  });

  it('errors when other and details empty', () => {
    cy.get('input[name="when-to-step-in"]').check('another-way', { force: true });
    cy.contains('button', 'Save and continue').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Enter details of when your replacement attorneys should step in');
    });

    cy.contains('.govuk-fieldset .govuk-error-message', 'Enter details of when your replacement attorneys should step in');
  });
});
