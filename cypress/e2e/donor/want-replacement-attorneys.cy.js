describe('Do you want replacement attorneys', () => {
  it('wants replacement attorneys - acting jointly', () => {
    cy.visit('/fixtures?redirect=/do-you-want-replacement-attorneys&progress=chooseYourAttorneys&attorneys=jointly');

    cy.get('div.govuk-warning-text').should('contain', 'Replacement attorneys are an important backup when attorneys are appointed to act jointly.')

    cy.get('input[name="yes-no"]').check('yes', { force: true })

    cy.checkA11yApp();

    cy.contains('button', 'Save and continue').click();
    cy.url().should('contain', '/choose-replacement-attorneys');
  });

  it('wants replacement attorneys - acting jointly for some and severally for others', () => {
    cy.visit('/fixtures?redirect=/do-you-want-replacement-attorneys&progress=chooseYourAttorneys&attorneys=jointly-for-some-severally-for-others');

    cy.get('div.govuk-warning-text').should('contain', 'You appointed your attorneys to act jointly for some decisions, and jointly and severally for others.')

    cy.get('input[name="yes-no"]').check('yes', { force: true })

    cy.checkA11yApp();

    cy.contains('button', 'Save and continue').click();
    cy.url().should('contain', '/choose-replacement-attorneys');
  });

  it('wants replacement attorneys - acting jointly and severally', () => {
    cy.visit('/fixtures?redirect=/do-you-want-replacement-attorneys&progress=chooseYourAttorneys');

    cy.get('div.govuk-warning-text').should('not.exist')

    cy.get('input[name="yes-no"]').check('yes', { force: true })

    cy.checkA11yApp();

    cy.contains('button', 'Save and continue').click();
    cy.url().should('contain', '/choose-replacement-attorneys');
  });

  it('does not want replacement attorneys', () => {
    cy.visit('/fixtures?redirect=/do-you-want-replacement-attorneys&progress=chooseYourAttorneys');

    cy.get('div.govuk-warning-text').should('not.exist')

    cy.get('input[name="yes-no"]').check('no', { force: true })

    cy.checkA11yApp();

    cy.contains('button', 'Save and continue').click();
    cy.url().should('contain', '/task-list');

    cy.contains('a', 'Choose your replacement attorneys').parent().parent().contains('Completed')
  });

  it('errors when unselected', () => {
    cy.visit('/fixtures?redirect=/do-you-want-replacement-attorneys&progress=chooseYourAttorneys');
    cy.contains('button', 'Save and continue').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Select yes to add replacement attorneys');
    });

    cy.contains('.govuk-fieldset .govuk-error-message', 'Select yes to add replacement attorneys');
  });
});
