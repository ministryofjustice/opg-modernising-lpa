describe('Upload evidence', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/upload-evidence');
    });

    it('can upload evidence', () => {
        cy.checkA11yApp();

        cy.get('input[type="file"]').attachFile(['dummy.pdf', 'dummy.png']);

        cy.contains('button', 'Upload files').click()

        cy.url().should('contain', '/upload-evidence');

        cy.checkA11yApp();

        cy.get('.govuk-notification-banner--success').within(() => {
            cy.contains('2 files successfully uploaded');
        });

        cy.get('.govuk-summary-list').within(() => {
            cy.contains('dummy.pdf');
            cy.contains('dummy.png');
        });

        cy.contains('button', 'Continue to payment').click()

        cy.url().should('contain', '/payment-confirmation');
    });
});
