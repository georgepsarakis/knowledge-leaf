import {
    Center, ChakraProvider, Container, defaultSystem, Image,
    Heading, Spinner, Text, Box, Button, Link, Float, ListRoot, ListItem
} from "@chakra-ui/react";
import { Swiper, SwiperSlide } from 'swiper/react';
import 'swiper/css';
import "./index.css";
import React, {useEffect, useState} from "react";
import {getDomain} from "./backend/backend";
import {Navigation} from 'swiper/modules';
import 'swiper/css/navigation';
import {AiOutlineArrowLeft, AiOutlineArrowRight} from "react-icons/ai";
import Layout from "./Layout";
import {LuExternalLink} from "react-icons/lu";

type ArticleImage = {
    url: string;
    height: number;
    width: number;
}

type OnThisDayEvent = {
    description: string;
    title: string;
    year: number;
    short_title: string;
    extract: string;
    image: ArticleImage;
    url: string;
    references: OnThisDayEventReference[];
}

type OnThisDayEventReference = {
    url: string;
    title: string;
}

type OnThisDayEventsResponse = {
    titles: OnThisDayEvent[];
}

function EventsOnThisDay() {
    const [data, setData] = useState<OnThisDayEventsResponse>();
    const [isLoading, setIsLoading] = useState<boolean>(true)
    let ignore = false;
    useEffect(() => {
        if (!ignore) {
            ignore = true;
            fetch(getDomain()+'/on-this-day/events').then(response => {
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
    return <ChakraProvider value={defaultSystem}>
        <Layout/>
        <Center>
                <Container marginTop={5} maxW={"4xl"}>
                    <Box>
                        <Float placement={"middle-start"}>
                            <Button backgroundColor="teal" className={"kl-swiper-button-prev"} size={"md"}><AiOutlineArrowLeft/></Button>
                        </Float>
                    </Box>
                    <Center>
                        <Heading size={"2xl"} color={"gray.500"}>On This Day</Heading>
                    </Center>
                            {
                                isLoading ?
                                    <Spinner size={"xl"} color={"teal.500"} /> :
                                    <Swiper
                                        modules={[Navigation]}
                                        navigation={{
                                            nextEl: '.kl-swiper-button-next',
                                            prevEl: '.kl-swiper-button-prev',
                                        }}
                                        slidesPerView={1}>
                                        {
                                            data!.titles.map((t) => {
                                                const key = t.url + "-" + t.year.toString();
                                                const dt = new Date();
                                                dt.setFullYear(t.year);
                                                const displayDate = dt.toLocaleDateString(
                                                    undefined,
                                                    {month: 'long', year: "numeric", day: "numeric"});
                                                const displayTitle = `${displayDate}: ${t.title}`;
                                                return <SwiperSlide key={t.url}>
                                                    <Box p="4"
                                                         borderWidth="1px"
                                                         borderColor="border.disabled"
                                                         color="fg.disabled" margin={30}>
                                                    <Event
                                                        key={key}
                                                        title={displayTitle}
                                                        extract={t.extract}
                                                        image={t.image}
                                                        short_title={t.short_title}
                                                        description={t.description}
                                                        url={t.url}
                                                        references={t.references}
                                                    />
                                                    </Box>
                                                </SwiperSlide>
                                            })
                                        }
                                    </Swiper>
                            }
                    <Float placement={"middle-end"}>
                        <Button backgroundColor="teal" className={"kl-swiper-button-next"} size={"md"}><AiOutlineArrowRight/></Button>
                    </Float>
                </Container>

        </Center>
       </ChakraProvider>;
}

type EventProps = {
    title: string;
    extract: string;
    image: ArticleImage;
    short_title: string;
    description: string;
    url: string;
    references: OnThisDayEventReference[];
}

function Event(props: EventProps) {
    const image = props.image;
    return <Container maxW={"2xl"}>
             <Heading size={"lg"} marginBottom={5}>{props.title}</Heading>
             <Center>
                <Image src={image.url}
                       width={image.width}
                       height={image.height}
                       transition={"0.5s linear"} />
             </Center>
             <Text marginTop={5} marginBottom={5}>{props.extract}</Text>
             <Link target={"_blank"} href={props.url}>
                <Text color={"teal"}>{props.short_title} - {props.description}</Text>
                 <LuExternalLink color={"teal"}/>
             </Link>
             <RenderReferences items={props.references}/>
           </Container>;
}

type ReferencesProps = {
    items: OnThisDayEventReference[];
}
function RenderReferences({ items }: ReferencesProps) {
    if (items === null || items.length === 0) {
        return null;
    }

    return <Box marginTop={2}>
        <Heading size={"md"} marginBottom={2}>References</Heading>
        <ListRoot>
            {items.map((ref) =>
                <ListItem key={ref.url}>
                    <Link target={"_blank"} href={ref.url}>
                        <Text color={"teal"}>{ref.title}</Text><LuExternalLink color={"teal"}/>
                    </Link>
                </ListItem>
            )}
        </ListRoot>
    </Box>
}

export default EventsOnThisDay;