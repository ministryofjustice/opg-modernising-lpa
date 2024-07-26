import { TestEmail } from "../../support/e2e";

describe('Choose replacement attorneys summary', () => {
  beforeEach(() => {
    cy.visit('/fixtures?redirect=/choose-replacement-attorneys-summary&progress=chooseYourReplacementAttorneys');
  });

  it('multiple attorneys details are listed', () => {
    cy.checkA11yApp();

    cy.contains('You have added 2 replacement attorneys');

    cy.contains('Blake Buckley');
    cy.contains('2 January 2000');
    cy.contains('2 RICHMOND PLACE');
    cy.contains('B14 7ED');

    cy.contains('Taylor Thompson');
    cy.contains('2 January 2000');
    cy.contains('2 RICHMOND PLACE');
    cy.contains('B14 7ED');

    cy.visitLpa('/task-list')
    cy.contains('a', 'Choose your replacement attorneys').parent().parent().contains('2 added')
  });

  it('can amend attorney details', () => {
    cy.checkA11yApp();

    cy.contains('.govuk-summary-card', 'Blake Buckley').contains('a', 'Change').click();
    cy.get('#f-first-names').clear().type('Mark');

    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/choose-replacement-attorneys-summary');
    cy.contains('Mark Buckley');
  });

  it('can amend attorney address', () => {
    cy.checkA11yApp();

    cy.contains('.govuk-summary-card', 'Taylor Thompson')
      .contains('.govuk-summary-list__row', 'Address')
      .contains('a', 'Change')
      .click();

    cy.get('#f-address-line-1').clear().type('4 RICHMOND PLACE');
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/choose-replacement-attorneys-summary');
    cy.contains('dd', '4 RICHMOND PLACE');
  });

  it('can add another attorney from summary page', () => {
    cy.checkA11yApp();

    cy.get('input[name="yes-no"]').check('yes', { force: true });
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/choose-replacement-attorneys');

    cy.get('#f-first-names').clear().type('Bob Arnold');
    cy.get('#f-last-name').clear().type('Jones');
    cy.get('#f-email').clear().type(TestEmail);
    cy.get('input[name="date-of-birth-day"]').clear().type('31');
    cy.get('input[name="date-of-birth-month"]').clear().type('12');
    cy.get('input[name="date-of-birth-year"]').clear().type('1995');
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/choose-replacement-attorneys-address');

    cy.contains('label', 'Enter a new address').click();
    cy.contains('button', 'Continue').click();

    cy.get('#f-lookup-postcode').type('B14 7ED');
    cy.contains('button', 'Find address').click();

    cy.get('#f-select-address').select('5 RICHMOND PLACE, BIRMINGHAM, B14 7ED');
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/choose-replacement-attorneys-address');
    cy.get('#f-address-line-1').should('have.value', '5 RICHMOND PLACE');
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/choose-replacement-attorneys-summary');

    cy.contains('Bob Arnold Jones');
    cy.contains('31 December 1995');
    cy.contains('5 RICHMOND PLACE');
    cy.contains('B14 7ED');
  });

  it('can remove an attorney', () => {
    cy.checkA11yApp();

    cy.contains('.govuk-summary-card', 'Blake Buckley').contains('a', 'Remove').click();

    cy.url().should('contain', '/remove-replacement-attorney');
    cy.url().should('match', /id=\w*/);

    cy.checkA11yApp();

    cy.contains('Are you sure you want to remove Blake Buckley?');

    cy.get('input[name="yes-no"]').check('yes', { force: true });
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/choose-replacement-attorneys-summary');

    cy.checkA11yApp();

    cy.get('main').should('not.contain', 'Blake Buckley');

    cy.contains('.govuk-summary-card', 'Taylor Thompson').contains('a', 'Remove').click();
    cy.get('input[name="yes-no"]').check('yes', { force: true });
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/do-you-want-replacement-attorneys');
  });

  it('errors when remove not selected', () => {
    cy.contains('.govuk-summary-card', 'Blake Buckley').contains('a', 'Remove').click();

    cy.contains('button', 'Continue').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Select yes to remove the replacement attorney');
    });

    cy.contains('.govuk-fieldset .govuk-error-message', 'Select yes to remove the replacement attorney');
  });
});
