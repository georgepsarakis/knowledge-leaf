import React from 'react';
import './App.css';
import {
    Badge, Box,
    Center,
    Container,
    Flex,
    Heading,
    Image,
    Link,
    List,
    Separator,
    Spinner,
} from "@chakra-ui/react";
import {Button} from "@chakra-ui/react";
import { ChakraProvider, defaultSystem } from "@chakra-ui/react"
import { Text } from "@chakra-ui/react"
import { useState, useEffect } from 'react';
import {HiAcademicCap} from "react-icons/hi2";
import {BiBarChart} from "react-icons/bi";

import {parse} from 'tldts';
import {Tag} from "./components/ui/tag";
import Layout from "./Layout";
import {getDomain} from "./backend/backend";
import {LuExternalLink} from "react-icons/lu";


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
    categories: string[];
}

const ArticleTypeDisambiguation = 'disambiguation';

type TriviaResponse = {
    results: Trivia[];
}

const wikiLogo = "https://upload.wikimedia.org/wikipedia/commons/6/63/Wikipedia-logo.png";

function DisplayRandomTrivia() {
    const [data, setData] = useState<TriviaResponse>({
        results: []
    })
    const [isLoading, setIsLoading] = useState<boolean>(true)
    let ignore = false;
    useEffect(() => {
        if (!ignore) {
            ignore = true;
            // TODO: prefetch next request
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
        <Container transition={"background 1s ease-in-out"} marginBottom={2}>
            { isLoading ?
                <Center>
                    <Box position="relative" aria-busy="true" userSelect="none">
                        <Container maxW={"3xl"} height={"lg"} animationName="fade-in" animationDuration="slow">
                                <Center>
                                    <Box pos="absolute" inset="0" bg="bg/80">
                                        <Center h={"full"}>
                                            <Spinner size={"xl"} color={"teal.500"}/>
                                        </Center>
                                    </Box>
                                </Center>
                        </Container>
                    </Box>
                </Center> :
                data.results.map((trivia) => (
                    <Container fluid key={trivia.summary} animationName="fade-in" animationDuration="slow">
                        <Center marginTop="5" marginBottom="10">
                            {trivia.metadata.image.url ?
                                <Image src={trivia.metadata.image.url}
                                       width={trivia.metadata.image.width}
                                       height={trivia.metadata.image.height}
                                       transition={"0.5s linear"} /> :
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
                            <Text fontWeight="semibold">{trivia.title}{trivia.metadata.description ? ' - ' + trivia.metadata.description : ''}</Text>
                            <LuExternalLink />
                        </Link>
                        {
                            trivia.categories &&
                            <div>
                                <Separator marginTop={5} marginBottom={5} />
                                {
                                    trivia.categories.map((item) => {
                                        return <Tag marginLeft={2} marginBottom={2} key={item} fontStyle={"italic"}
                                                    size={"lg"} colorPalette={"orange"}>
                                            <Link href={"https://en.wikipedia.org/wiki/Category:"+encodeURIComponent(item)} target={"_blank"}>
                                                {item}
                                            </Link>
                                        </Tag>
                                    })
                                }
                            </div>
                        }
                    </Container>
                ))
            }
        </Container>
    );
}

// TODO: add footer (About links etc)
function App() {
    const [triviaID, setTriviaID] = useState<number>(1);

    const newRandomTrivia = () => {
        return <DisplayRandomTrivia key={triviaID} />;
    };
    return <ChakraProvider value={defaultSystem}>
        <Layout/>
        <Center>
            <Container maxW={"3xl"}>
                { triviaID > 0 &&
                    <Container marginTop={2} fluid>
                        <Center>
                            <Flex gap={4}>
                            <Text textStyle="xl">Today's Reading</Text>
                            <Badge variant="solid" colorPalette="teal">
                                <BiBarChart size={20} />
                                <Text textStyle="xl"><b>{triviaID}</b></Text>
                            </Badge>
                            </Flex>
                        </Center>
                    </Container>
                }
                { newRandomTrivia() }
                <Center>
                    <Container marginTop={2} maxW={"3xl"}>
                        <Center>
                            <Button backgroundColor="teal" size={"2xl"} w={"100%"}
                                    onClick={() => {
                                        setTriviaID(triviaID => triviaID + 1);
                                    }}>
                                <HiAcademicCap/>
                                Random Trivia
                            </Button>
                        </Center>
                    </Container>
                </Center>
            </Container>
        </Center>
    </ChakraProvider>;
}

export default App;
