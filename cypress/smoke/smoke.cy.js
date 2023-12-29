describe('Smoke tests', () => {
    describe('external dependencies', () => {
        describe('UID service', () => {
            it('request signing and base URL are configured correctly', () => {
                cy.request('/health-check/dependency').should((response) => {
                    expect(response.status).not.to.eq(403)
                })
            })
        })
    })

    describe('app', () => {
        it('is available', () => {
            cy.visit('/')

            cy.get('h1').should('contain', 'Make and register a lasting power of attorney (LPA)');
            cy.contains('a', 'Start');
        })
    })
})
