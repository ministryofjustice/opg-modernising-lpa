describe('Dev features', () => {

    it('Show translation keys', () => {
        cy.visit('/fixtures?redirect=/when-can-the-lpa-be-used?showTranslationKeys=1&progress=chooseYourAttorneys');
        cy.get('h1').should(
            "have.text",
            "{When your attorneys can use your LPA} [whenYourAttorneysCanUseYourLpa]"
        );
        cy.contains('a', 'Toggle translation keys').click();

        cy.url().should('contain', '/when-can-the-lpa-be-used');
        cy.get('h1').should(
            "have.text",
            "When your attorneys can use your LPA"
        );
    });
});
