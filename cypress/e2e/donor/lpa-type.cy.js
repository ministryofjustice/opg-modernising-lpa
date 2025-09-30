import { AddressFormAssertions, eventLoggerUrl } from "../../support/e2e";

describe('LPA type', () => {
    it('can be submitted', () => {
        cy.visit('/fixtures?redirect=/your-name');

        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Doe');

        cy.contains('button', 'Save and continue').click();

        cy.get('#f-date-of-birth').invoke('val', '1');
        cy.get('#f-date-of-birth-month').invoke('val', '2');
        cy.get('#f-date-of-birth-year').invoke('val', '1990');

        cy.contains('button', 'Save and continue').click();

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        AddressFormAssertions.assertCanAddAddressFromSelect()

        cy.contains('a', 'Continue').click();

        cy.get('[name="selected"]').check('Yes', { force: true })

        cy.contains('button', 'Save and continue').click();

        cy.get('[name="contact-language"]').check('en', { force: true })
        cy.get('[name="lpa-language"]').check('en', { force: true })

        cy.contains('button', 'Save and continue').click()

        cy.contains('a', 'Continue').click();

        cy.checkA11yApp();

        cy.get('[name="lpa-type"]').check('property-and-affairs', { force: true })

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/task-list');
        cy.waitForTextVisibilityByReloading('.govuk-summary-list__value', 'M-');

        cy.url().then((url) => {
            cy.origin(eventLoggerUrl(), { args: { url } }, ({ url }) => {
                cy.visit(`/?detail-type=uid-requested&detail=${url.split('/')[4]}`);
                cy.contains(`"lpaID":"${url.split('/')[4]}"`);
            });
        });

        cy.visit('/dashboard')

        cy.contains('.govuk-body-s', 'Reference number:')
            .invoke('text')
            .then((text) => {
                const uid = text.split(':')[1].trim();

                cy.origin(eventLoggerUrl(), { args: { uid } }, ({ uid }) => {
                    cy.visit(`/?detail-type=application-updated&detail=${uid}`);
                    cy.contains(`"uid":"${uid}"`);
                    cy.contains('"type":"property-and-affairs"');
                });
            });
    });

    it('errors when unselected', () => {
        cy.visit('/fixtures?redirect=/lpa-type');

        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select the type of LPA to make');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select the type of LPA to make');
    });
});
