import { TestEmail } from "../../support/e2e";

describe('Choose attorneys', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/choose-attorneys');
    });

    it('can be submitted', () => {
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-email').type(TestEmail);
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');

        cy.checkA11yApp();

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-attorneys-address');

        cy.contains("John Doe’s address");
    });

    it('can choose not to submit email', () => {
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');

        cy.checkA11yApp();

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-attorneys-address');

        cy.contains("John Doe’s address");
    });

    it('errors when empty', () => {
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter first names');
            cy.contains('Enter last name');
            cy.contains('Enter date of birth');
        });

        cy.contains('[for=f-first-names] + .govuk-error-message', 'Enter first names');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Enter date of birth');
    });

    it('errors when names too long', () => {
        cy.get('#f-first-names').invoke('val', 'a'.repeat(54));
        cy.get('#f-last-name').invoke('val', 'b'.repeat(62));

        cy.contains('button', 'Save and continue').click();

        cy.contains('[for=f-first-names] + .govuk-error-message', 'First names must be 53 characters or less');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
    });

    it('errors when invalid email', () => {
        cy.get('#f-email').type('not-an-email');

        cy.contains('button', 'Save and continue').click();

        cy.contains('[for=f-email] + .govuk-error-message', 'Email address must be in the correct format, like name@example.com');
    });

    it('errors when invalid dates of birth', () => {
        cy.get('#f-date-of-birth').type('1');
        cy.contains('button', 'Save and continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must include a month and year');

        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('2222');
        cy.contains('button', 'Save and continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be in the past');

        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').clear().type('1990');
        cy.contains('button', 'Save and continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be a real date');
    });

    it('warns when name shared with other actor', () => {
        cy.visit('/fixtures?redirect=/choose-attorneys&progress=provideYourDetails');

        cy.get('#f-first-names').type('Sam');
        cy.get('#f-last-name').type('Smith');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-attorneys');

        cy.contains('The donor’s name is also Sam Smith.');

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-attorneys-address');
    });
});
