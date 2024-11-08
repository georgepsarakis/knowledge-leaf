import React from 'react';
import './App.css';
import {
    Badge,
    Center,
    Container,
    Heading, HStack,
    Image,
    Link,
    List,
    Separator,
    Spinner,
    Stack
} from "@chakra-ui/react"
import {Button} from "@chakra-ui/react";
import { ChakraProvider, defaultSystem } from "@chakra-ui/react"
import { Text } from "@chakra-ui/react"
import { useState, useEffect } from 'react';
import {HiAcademicCap} from "react-icons/hi2";
import {BiBarChart} from "react-icons/bi";

import {parse} from 'tldts';

type TriviaImage = {
    url: string;
    height: number;
    width: number;
}

type TriviaMetadata = {
    description: string;
    url: string;
    image: TriviaImage;
}

type Trivia = {
    title: string;
    summary: string;
    type: string;
    metadata: TriviaMetadata;
}

const ArticleTypeDisambiguation = 'disambiguation';

type TriviaResponse = {
    results: Trivia[];
}

const wikiLogo = "https://upload.wikimedia.org/wikipedia/commons/6/63/Wikipedia-logo.png";

function getDomain() {
    const domain = window.location.hostname;
    if (domain == "localhost") {
        return 'http://' + domain + ":4000";
    }
    const tld = parse(window.location.hostname)
    return 'https://api.' + tld.domain;
}

function DisplayRandomTrivia() {
    const [data, setData] = useState<TriviaResponse>({
        results: []
    })
    const [isLoading, setIsLoading] = useState<boolean>(true)
    let ignore = false;
    useEffect(() => {
        if (!ignore) {
            ignore = true;
            fetch(getDomain()+'/trivia/random').then(response => {
                return response.json()
            }).then(triviaData => {
                setData(triviaData);
                setIsLoading(false);
            }).catch((error) => {
                console.log(error);
            });
        }
        return () => {
            ignore = true;
        }
    }, []);

    return (
        <Container transition={"background 0.2s ease-in-out"}>
            { isLoading ? <Center><Spinner/></Center> :
                data.results.map((trivia) => (
                    <Container fluid key={trivia.summary}>
                        <Center marginTop="5" marginBottom="10">
                            {trivia.metadata.image.url ?
                                <Image src={trivia.metadata.image.url}
                                       width={trivia.metadata.image.width}
                                       height={trivia.metadata.image.height}
                                       transition={"0.8s linear"} /> :
                                <Image src={wikiLogo}/>}
                        </Center>
                        <Heading size={"lg"}>{trivia.title}
                            { trivia.type == ArticleTypeDisambiguation ? ` (${trivia.type})` : '' }
                        </Heading>
                        { trivia.type != ArticleTypeDisambiguation ?
                            <Text style={{whiteSpace: "pre-wrap"}}>{trivia.summary}</Text> :
                            <List.Root>
                                {
                                    trivia.summary.split("\n").map((item) => {
                                        return <List.Item key={item}>{item}</List.Item>
                                    })
                                }
                            </List.Root>
                        }

                        <Separator marginTop={5} marginBottom={5} />

                        <Link color={"teal.500"} href={trivia.metadata.url} target={"_blank"}>
                            <Text fontStyle="italic">{trivia.title}{trivia.metadata.description ? ' - ' + trivia.metadata.description : ''}</Text>
                        </Link>
                    </Container>
                ))
            }
        </Container>
    );
}

function App() {
    const [triviaID, setTriviaID] = useState<number>(0)
    return <ChakraProvider value={defaultSystem}>
        <Center>
            <Container marginTop={10} maxW={"3xl"}>
                <Stack>
                    <Button backgroundColor="teal"
                            onClick={() => { setTriviaID(triviaID => triviaID + 1) }}>
                        <HiAcademicCap/>
                        Random Trivia
                    </Button>
                    { (triviaID > 0) && <DisplayRandomTrivia key={triviaID} /> }
                    { triviaID > 0 &&
                        <Container marginTop={5}>
                        <Separator marginTop={5} marginBottom={10}/>
                            <HStack>
                                <Text textStyle="xl">Today's Reading</Text>
                                <Badge variant="solid" colorPalette="teal">
                                    <BiBarChart size={20} />
                                    <Text textStyle="xl"><b>{triviaID}</b></Text>
                                </Badge>
                            </HStack>
                        </Container>
                    }
                </Stack>
            </Container>
        </Center>

    </ChakraProvider>;
}

export default App;
