import {AttorneyNames, CertificateProviderName, CorrespondentName, PeopleToNotifyNames} from "../../support/e2e.js";

describe('Your name', () => {
    describe('first time', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/your-name');
        });

        it('can be submitted', () => {
            cy.checkA11yApp();
            cy.contains('a', 'Return to task list').should('not.exist');

            cy.get('#f-first-names').invoke('val', 'John');
            cy.get('#f-last-name').invoke('val', 'Doe');

            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-date-of-birth');
        });

        it('errors when empty', () => {
            cy.contains('button', 'Save and continue').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Enter first names');
                cy.contains('Enter last name');
            });

            cy.contains('[for=f-first-names] + div + .govuk-error-message', 'Enter first names');
            cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
        });

        it('errors when names too long', () => {
            cy.get('#f-first-names').invoke('val', 'a'.repeat(54));
            cy.get('#f-last-name').invoke('val', 'b'.repeat(62));
            cy.get('#f-other-names').invoke('val', 'c'.repeat(51));

            cy.contains('button', 'Save and continue').click();

            cy.contains('[for=f-first-names] + div + .govuk-error-message', 'First names must be 53 characters or less');
            cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
            cy.contains('[for=f-other-names] + div + .govuk-error-message', 'Other names you are known by must be 50 characters or less');
        });
    });

    describe('after completing', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/your-name&progress=chooseYourAttorneys');
        });

        it('shows task list button', () => {
            cy.contains('a', 'Return to task list');
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-date-of-birth');
        });

        it('warns when name shared with other actor', () => {
            cy.visit('/fixtures?redirect=/task-list&progress=peopleToNotifyAboutYourLpa');

            cy.contains('li', 'Add a correspondent').should('contain', 'Not started').click();

            cy.checkA11yApp();
            cy.contains('label', 'Yes').click();
            cy.contains('button', 'Save and continue').click();

            cy.checkA11yApp();
            cy.get('#f-first-names').invoke('val', CorrespondentName.FirstNames);
            cy.get('#f-last-name').invoke('val', CorrespondentName.LastName);
            cy.get('#f-email').invoke('val', 'email@example.com');
            cy.contains('label', 'No').click();
            cy.contains('button', 'Save and continue').click();

            cy.visitLpa("/your-name")
            cy.get('#f-first-names').invoke('val', AttorneyNames[0].FirstNames);
            cy.get('#f-last-name').invoke('val', AttorneyNames[0].LastName);
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/warning');

            cy.contains('You have the same name as your attorney. As the donor, you cannot act as an attorney for your LPA.');
            cy.contains('dt', 'First names').parent().contains('a', 'Change').click();

            changeNameTo(cy, CertificateProviderName.FirstNames, CertificateProviderName.LastName)

            cy.contains('You have the same name or address as your certificate provider. As the donor, you cannot act as the certificate provider for your LPA.');
            cy.contains('dt', 'First names').parent().contains('a', 'Change').click();

            changeNameTo(cy, CorrespondentName.FirstNames, CorrespondentName.LastName)

            cy.contains('You and your correspondent have the same name. As the donor, you will automatically receive correspondence from the Office of the Public Guardian unless you nominate another person for this role.');
            cy.contains('dt', 'First names').parent().contains('a', 'Change').click();

            changeNameTo(cy, PeopleToNotifyNames[0].FirstNames, PeopleToNotifyNames[0].LastName)

            cy.contains('You and your person to notify have the same name. As the donor, you will automatically receive updates from the Office of the Public Guardian – you do not need to be a person to notify.');
            cy.contains('a', 'Continue').click();
            cy.url().should('contain', '/your-date-of-birth');
        });

        function changeNameTo(cy, firstNames, lastNames) {
            cy.url().should('contain', '/your-name');
            cy.get('#f-first-names').invoke('val', firstNames);
            cy.get('#f-last-name').invoke('val', lastNames);
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/warning');
        }
    });
});
