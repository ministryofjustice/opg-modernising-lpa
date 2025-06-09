describe('Read the LPA', () => {
    it('displays the LPA details with actor specific content', () => {
        cy.visit('/fixtures/attorney?redirect=/read-the-lpa');

        cy.contains('Donor: Sam Smith');
        cy.contains('Certificate provider: Charlie Cooper');
        cy.contains('Attorney: Jessie Jones');
        cy.contains('Trust corporation attorney: First Choice Trust Corporation Ltd.');
        cy.contains('Replacement attorney: Blake Buckley');
        cy.contains('Replacement trust corporation attorney: Second Choice Trust Corporation Ltd.');
        cy.contains('Signed by Sam Smith on: 2 January 2023');
        cy.contains('Witnessed by Charlie Cooper on: 2 January 2023');
        cy.contains('Signed by Charlie Cooper on: 2 January 2023');

        cy.contains('Continue').click();

        cy.url().should('contain', '/task-list');
    });

    it('redirects to registration language if on task list in other language', () => {
        cy.visit('/fixtures/attorney?redirect=/task-list');
        cy.contains('a', 'Cymraeg').click();
        cy.contains('a', 'Darllen yr LPA').click();

        cy.contains('h1', 'Read the LPA');
    });
});
