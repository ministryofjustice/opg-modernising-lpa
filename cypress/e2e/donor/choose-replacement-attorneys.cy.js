import {DateOfBirthAssertions, TestEmail} from "../../support/e2e";

describe('Choose replacement attorneys', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/do-you-want-replacement-attorneys&progress=chooseYourAttorneys');

        cy.get('input[name="yes-no"]').check('yes', { force: true })
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/choose-replacement-attorneys');
    });

    it('can be submitted', () => {
        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Doe');
        cy.get('#f-email').invoke('val', TestEmail);
        cy.get('#f-date-of-birth').invoke('val', '1');
        cy.get('#f-date-of-birth-month').invoke('val', '2');
        cy.get('#f-date-of-birth-year').invoke('val', '1990');

        cy.checkA11yApp();

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys-address');
        cy.contains("John Doe’s address");
    });

    it('can choose not to submit email', () => {
        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Doe');
        cy.get('#f-date-of-birth').invoke('val', '1');
        cy.get('#f-date-of-birth-month').invoke('val', '2');
        cy.get('#f-date-of-birth-year').invoke('val', '1990');

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
        cy.get('#f-email').invoke('val', 'not-an-email');

        cy.contains('button', 'Save and continue').click();

        cy.contains('[for=f-email] + .govuk-error-message', 'Email address must be in the correct format, like name@example.com');
    });

    it('errors when invalid dates of birth', () => {
        DateOfBirthAssertions.assertInvalidDatesOfBirth()
    });

    it('warns when name shared with other actor', () => {
        cy.get('#f-first-names').invoke('val', 'Sam');
        cy.get('#f-last-name').invoke('val', 'Smith');
        cy.get('#f-date-of-birth').invoke('val', '1');
        cy.get('#f-date-of-birth-month').invoke('val', '2');
        cy.get('#f-date-of-birth-year').invoke('val', '1990');
        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');

        cy.contains('You and your replacement attorney have the same name. As the donor, you cannot act as a replacement attorney for your LPA.');

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys-address');
    });

    it('permanently warns when date of birth is under 18', () => {
        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Doe');
        cy.get('#f-date-of-birth').invoke('val', '1');
        cy.get('#f-date-of-birth-month').invoke('val', '2');
        cy.get('#f-date-of-birth-year').invoke('val', new Date().getFullYear() - 1);
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
        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Doe');
        cy.get('#f-date-of-birth').invoke('val', '1');
        cy.get('#f-date-of-birth-month').invoke('val', '2');
        cy.get('#f-date-of-birth-year').invoke('val', '1900');
        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');

        cy.contains('By continuing, you confirm that this person is more than 100 years old. If not, please change their date of birth.');

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys-address');

        cy.go(-2);
        cy.url().should('contain', '/choose-replacement-attorneys');

        cy.get('#f-date-of-birth-year').invoke('val', new Date().getFullYear() - 20);
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/choose-replacement-attorneys-address');
    });
});
