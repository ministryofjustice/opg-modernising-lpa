import {
    AttorneyNames,
    DateOfBirthAssertions,
    DonorName,
    PeopleToNotifyNames,
    ReplacementAttorneyNames,
    TestEmail
} from "../../support/e2e";

describe('Enter attorney', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/choose-attorneys-guidance&progress=provideYourDetails');
        cy.contains('a', 'Continue').click();
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
        cy.url().should('contain', '/choose-attorneys-address');

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
        cy.visit('/fixtures?redirect=/choose-attorneys-summary&progress=peopleToNotifyAboutYourLpa');

        cy.contains('dd', `${AttorneyNames[1].FirstNames} ${AttorneyNames[1].LastName}`).parent().contains('a', 'Change').click();
        changeNameTo(cy, DonorName.FirstNames, DonorName.LastName)

        cy.contains('You and your attorney have the same name. As the donor, you cannot act as an attorney for your LPA.');

        cy.contains('dt', 'First names').parent().contains('a', 'Change').click();
        changeNameTo(cy, AttorneyNames[0].FirstNames, AttorneyNames[0].LastName)

        cy.contains(`You have already entered ${AttorneyNames[0].FirstNames} ${AttorneyNames[0].LastName} as an attorney on your LPA.`);

        cy.contains('dt', 'First names').parent().contains('a', 'Change').click();
        changeNameTo(cy, PeopleToNotifyNames[0].FirstNames, PeopleToNotifyNames[0].LastName)

        cy.contains(`${PeopleToNotifyNames[0].FirstNames} ${PeopleToNotifyNames[0].LastName} has the same name as another person you’ve chosen to act in this LPA`);

        cy.contains('dt', 'First names').parent().contains('a', 'Change').click();
        changeNameTo(cy, ReplacementAttorneyNames[0].FirstNames, ReplacementAttorneyNames[0].LastName)

        cy.contains(`${ReplacementAttorneyNames[0].FirstNames} ${ReplacementAttorneyNames[0].LastName} has the same name as another person you’ve chosen to act in this LPA`);

        cy.contains('a', 'Continue').click();
        cy.url().should('contain', '/choose-attorneys-address');
    });

    function changeNameTo(cy, firstNames, lastNames) {
        cy.url().should('contain', '/enter-attorney');
        cy.get('#f-first-names').invoke('val', firstNames);
        cy.get('#f-last-name').invoke('val', lastNames);
        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/warning');
    }

    it('permanently warns when date of birth is under 18', () => {
        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Doe');
        cy.get('#f-date-of-birth').invoke('val', '1');
        cy.get('#f-date-of-birth-month').invoke('val', '2');
        cy.get('#f-date-of-birth-year').invoke('val', new Date().getFullYear() - 1);
        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/warning');

        cy.contains('John Doe is under 18 years old. You can continue making your LPA but you will not be able to sign it until they are 18.');

        cy.contains('dt', 'Date of birth').parent().contains('a', 'Change').click();
        cy.url().should('contain', '/enter-attorney');

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/warning');

        cy.contains('John Doe is under 18 years old. You can continue making your LPA but you will not be able to sign it until they are 18.');
    });

    it('warns when date of birth is over 100', () => {
        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Doe');
        cy.get('#f-date-of-birth').invoke('val', '1');
        cy.get('#f-date-of-birth-month').invoke('val', '2');
        cy.get('#f-date-of-birth-year').invoke('val', '1900');
        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/warning');

        cy.contains('John Doe is over 100 years old.');

        cy.contains('a', 'Change date of birth for Sam Smith').click();

        cy.get('#f-date-of-birth-year').invoke('val', new Date().getFullYear() - 20);
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/choose-attorneys-address');
    });
});
