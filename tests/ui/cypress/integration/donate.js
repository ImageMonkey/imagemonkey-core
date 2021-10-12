// npm install --save-dev cypress-file-upload
// start with: node_modules/.bin/cypress open --project tests/ui/

describe('Donate Image', () => {
    before(() => {
        cy.clear_db();
		//cy.exec('cd ../ && go test -run TestDatabaseEmpty')
    })

    it('Testing image donation', () => {
        cy.donate_image('apple1.jpeg');
		/*cy.visit('http://127.0.0.1:8080/donate');

        cy.fixture('images/apples/apple1.jpeg').then(fileContent => {
            cy.get('[id="dropzone"]').attachFile({
                fileContent: fileContent.toString(),
                fileName: 'apple1.jpeg',
                mimeType: 'image/png'
            });
        });*/
    });
})
