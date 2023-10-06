describe('Upload evidence', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/upload-evidence&lpa.complete=1&lpa.certificateProvider=1');
    });

    it('can upload evidence', () => {
        cy.checkA11yApp();

        cy.get('input[type="file"]').attachFile(['example.pdf', 'another-example.pdf']);

        cy.contains('button', 'Upload files').click()

        cy.url().should('contain', '/upload-evidence');

        cy.checkA11yApp();

        cy.get('.govuk-notification-banner--success').within(() => {
            cy.contains('2 files successfully uploaded');
        });

        cy.get('.govuk-summary-list').within(() => {
            cy.contains('example.pdf');
            cy.contains('another-example.pdf');
        });

        cy.contains('button', 'Continue to payment').click()

        cy.url().should('contain', '/payment-confirmation');
    });
});
