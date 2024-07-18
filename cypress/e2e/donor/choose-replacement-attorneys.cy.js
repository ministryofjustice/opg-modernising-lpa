import { TestEmail } from "../../support/e2e";

describe('Choose replacement attorneys', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/choose-replacement-attorneys&progress=chooseYourAttorneys');
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
        cy.url().should('contain', '/choose-replacement-attorneys-address');
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
        cy.url().should('contain', '/choose-replacement-attorneys-address');

        cy.contains("John Doe’s address");
    });

    it('errors when empty', () => {
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter first names');
            cy.contains('Enter last name');
            cy.contains('Enter date of birth');
        });

        cy.contains('[for=f-first-names] + div + .govuk-error-message', 'Enter first names');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Enter date of birth');
    });

    it('errors when names too long', () => {
        cy.get('#f-first-names').invoke('val', 'a'.repeat(54));
        cy.get('#f-last-name').invoke('val', 'b'.repeat(62));

        cy.contains('button', 'Save and continue').click();

        cy.contains('[for=f-first-names] + div + .govuk-error-message', 'First names must be 53 characters or less');
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
        cy.visit('/fixtures?redirect=/choose-replacement-attorneys&progress=chooseYourAttorneys');

        cy.get('#f-first-names').type('Sam');
        cy.get('#f-last-name').type('Smith');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');

        cy.contains('The donor’s name is also Sam Smith. The donor cannot also be a replacement attorney. By saving this section, you are confirming that these are two different people with the same name.');

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys-address');
    });

    it('permanently warns when date of birth is under 18', () => {
        cy.visit('/fixtures?redirect=/choose-replacement-attorneys&progress=chooseYourAttorneys');

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type(new Date().getFullYear() - 1);
        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');

        cy.contains('This attorney is under 18 years old. You can continue making your LPA but you will not be able to sign it until they are 18.');

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys-address');

        cy.visitLpa("/choose-replacement-attorneys-summary")
        cy.contains('a', 'Change').click()
        cy.url().should('contain', '/choose-replacement-attorneys');

        cy.contains('This attorney is under 18 years old. You can continue making your LPA but you will not be able to sign it until they are 18.');
    });

    it('warns when date of birth is over 100', () => {
        cy.visit('/fixtures?redirect=/choose-replacement-attorneys&progress=chooseYourAttorneys');

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1900');
        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');

        cy.contains('By continuing, you confirm that this person is more than 100 years old. If not, please change their date of birth.');

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys-address');
    });
});
