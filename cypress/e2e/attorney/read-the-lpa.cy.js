describe('Read the LPA', () => {
  it('displays the LPA details with actor specific content', () => {
    cy.visit('/fixtures/attorney?redirect=/read-the-lpa');

    cy.contains('dt', "When attorneys can use the LPA")
    cy.contains('dt', "Attorney names")
    cy.contains('dt', "Replacement attorney names")

    cy.contains('Continue').click();

    cy.url().should('contain', '/task-list');
  });
});
