describe('LPA type', () => {
  it('can be submitted', () => {
    cy.visit('/fixtures?redirect=/lpa-type&progress=provideYourDetails');

    cy.get('#f-lpa-type').check('property-and-affairs');

    cy.checkA11yApp();

    cy.contains('button', 'Continue').click();
    cy.url().should('contain', '/task-list');

    cy.url().then((url) => {
      cy.visit(`http://localhost:9001/?detail-type=uid-requested&detail=${url.split('/')[4]}`);
      cy.contains(`"LpaID":"${url.split('/')[4]}"`);
    });
  });

  it('errors when unselected', () => {
    cy.visit('/fixtures?redirect=/lpa-type');

    cy.contains('button', 'Continue').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Select the type of LPA to make');
    });

    cy.contains('.govuk-fieldset .govuk-error-message', 'Select the type of LPA to make');
  });
});
